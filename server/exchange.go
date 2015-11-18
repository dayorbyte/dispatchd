package main

import (
	"encoding/json"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/interfaces"
	"github.com/jeffjenkins/mq/msgstore"
	"github.com/jeffjenkins/mq/queue"
	"sync"
	"time"
)

type extype uint8

const (
	EX_TYPE_DIRECT  extype = 1
	EX_TYPE_FANOUT  extype = 2
	EX_TYPE_TOPIC   extype = 3
	EX_TYPE_HEADERS extype = 4
)

type Exchange struct {
	name         string
	extype       extype
	durable      bool
	autodelete   bool
	internal     bool
	arguments    *amqp.Table
	system       bool
	bindings     []*Binding
	bindingsLock sync.Mutex
	incoming     chan amqp.Frame
	closed       bool
	deleteActive time.Time
	deleteChan   chan *Exchange
	msgStore     *msgstore.MessageStore
}

func (exchange *Exchange) close() {
	exchange.closed = true
}

func (exchange *Exchange) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":     exchangeTypeToName(exchange.extype),
		"bindings": exchange.bindings,
	})
}

func equivalentExchanges(ex1 *Exchange, ex2 *Exchange) bool {
	// NOTE: auto-delete is ignored for existing exchanges, so we
	// do not check it here.
	if ex1.name != ex2.name {
		return false
	}
	if ex1.extype != ex2.extype {
		return false
	}
	if ex1.durable != ex2.durable {
		return false
	}
	if ex1.internal != ex2.internal {
		return false
	}
	if !amqp.EquivalentTables(ex1.arguments, ex2.arguments) {
		return false
	}
	return true
}

func exchangeNameToType(et string) (extype, error) {
	switch {
	case et == "direct":
		return EX_TYPE_DIRECT, nil
	case et == "fanout":
		return EX_TYPE_FANOUT, nil
	case et == "topic":
		return EX_TYPE_TOPIC, nil
	case et == "headers":
		return EX_TYPE_HEADERS, nil
	default:
		return 0, fmt.Errorf("Unknown exchang type '%s', %d %d", et, len(et), len("direct"))
	}
}

func exchangeTypeToName(et extype) string {
	switch {
	case et == EX_TYPE_DIRECT:
		return "direct"
	case et == EX_TYPE_FANOUT:
		return "fanout"
	case et == EX_TYPE_TOPIC:
		return "topic"
	case et == EX_TYPE_HEADERS:
		return "headers"
	default:
		panic(fmt.Sprintf("bad exchange type: %d", et))
	}
}

func (exchange *Exchange) queuesForPublish(server *Server, msg *amqp.Message) map[string]*queue.Queue {
	var queues = make(map[string]*queue.Queue)
	switch {
	case exchange.extype == EX_TYPE_DIRECT:
		for _, binding := range exchange.bindings {
			if binding.matchDirect(msg.Method) {
				var _, alreadySeen = queues[binding.queueName]
				if alreadySeen {
					continue
				}
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queues[binding.queueName] = queue
			}
		}
	case exchange.extype == EX_TYPE_FANOUT:
		for _, binding := range exchange.bindings {
			if binding.matchFanout(msg.Method) {
				var _, alreadySeen = queues[binding.queueName]
				if alreadySeen {
					continue
				}
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queues[binding.queueName] = queue
			}
		}
	case exchange.extype == EX_TYPE_TOPIC:
		for _, binding := range exchange.bindings {
			if binding.matchTopic(msg.Method) {
				var _, alreadySeen = queues[binding.queueName]
				if alreadySeen {
					continue
				}
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queues[binding.queueName] = queue
			}
		}
	case exchange.extype == EX_TYPE_HEADERS:
		// TODO: implement
		panic("Headers is not implemented!")
	default:
		// TODO: can this happen? Seems like checks should be earlier
		panic("unknown exchange type!")
	}
	return queues
}

func (exchange *Exchange) publish(server *Server, msg *amqp.Message) (*amqp.BasicReturn, *AMQPError) {
	// Concurrency note: Since there is no lock we can, technically, have messages
	// published after the exchange has been closed. These couldn't be on the same
	// channel as the close is happening on, so that seems justifiable.
	if exchange.closed {
		if msg.Method.Mandatory || msg.Method.Immediate {
			var rm = exchange.returnMessage(msg, 313, "Exchange closed, cannot route to queues or consumers")
			return rm, nil
		}
		return nil, nil
	}
	queues := exchange.queuesForPublish(server, msg)

	if len(queues) == 0 {
		// If we got here the message was unroutable.
		if msg.Method.Mandatory || msg.Method.Immediate {
			var rm = exchange.returnMessage(msg, 313, "No queues available")
			return rm, nil
		}
	}

	var queueNames = make([]string, 0, len(queues))
	for k, _ := range queues {
		queueNames = append(queueNames, k)
	}

	// Immediate messages
	if msg.Method.Immediate {
		var consumed = false
		// Add message to message store
		queueMessagesByQueue, err := exchange.msgStore.AddMessage(msg, queueNames)
		if err != nil {
			return nil, NewSoftError(500, err.Error(), 60, 40)
		}
		// Try to immediately consumed it
		for name, queue := range queues {
			qms := queueMessagesByQueue[name]
			for _, qm := range qms {
				var oneConsumed = queue.ConsumeImmediate(qm)
				var rhs = make([]interfaces.MessageResourceHolder, 0)
				if !oneConsumed {
					exchange.msgStore.RemoveRef(qm, name, rhs)
				}
				consumed = oneConsumed || consumed
			}
		}
		if !consumed {
			var rm = exchange.returnMessage(msg, 313, "No consumers available for immediate message")
			return rm, nil
		}
		return nil, nil
	}

	// Add the message to the message store along with the queues we're about to add it to
	queueMessagesByQueue, err := exchange.msgStore.AddMessage(msg, queueNames)
	if err != nil {
		return nil, NewSoftError(500, err.Error(), 60, 40)
	}

	for name, queue := range queues {
		qms := queueMessagesByQueue[name]
		for _, qm := range qms {
			if !queue.Add(qm) {
				// If we couldn't add it means the queue is closed and we should
				// remove the ref from the message store. The queue being closed means
				// it is going away, so worst case if the server dies we have to process
				// and discard the message on boot.
				var rhs = make([]interfaces.MessageResourceHolder, 0)
				exchange.msgStore.RemoveRef(qm, name, rhs)
			}
		}
	}
	return nil, nil
}

func (exchange *Exchange) returnMessage(msg *amqp.Message, code uint16, text string) *amqp.BasicReturn {
	return &amqp.BasicReturn{
		Exchange:   exchange.name,
		RoutingKey: msg.Method.RoutingKey,
		ReplyCode:  code,
		ReplyText:  text,
	}
}

func (exchange *Exchange) addBinding(method *amqp.QueueBind, connId int64, fromDisk bool) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	var binding = NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments)

	for _, b := range exchange.bindings {
		if binding.Equals(b) {
			return nil
		}
	}

	if exchange.autodelete {
		exchange.deleteActive = time.Unix(0, 0)
	}
	exchange.bindings = append(exchange.bindings, binding)
	return nil
}

func (exchange *Exchange) bindingsForQueue(queueName string) []*Binding {
	var ret = make([]*Binding, 0)
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	for _, b := range exchange.bindings {
		if b.queueName == queueName {
			ret = append(ret, b)
		}
	}
	return ret
}

func (exchange *Exchange) removeBindingsForQueue(queueName string) {
	var remaining = make([]*Binding, 0)
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	for _, b := range exchange.bindings {
		if b.queueName != queueName {
			remaining = append(remaining, b)
		}
	}
	exchange.bindings = remaining
}

func (exchange *Exchange) removeBinding(server *Server, queue *queue.Queue, binding *Binding) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	// First de-persist
	if queue.Durable && exchange.durable {
		err := server.depersistBinding(binding)
		if err != nil {
			return err
		}
	}

	// Delete binding
	for i, b := range exchange.bindings {
		if binding.Equals(b) {
			exchange.bindings = append(exchange.bindings[:i], exchange.bindings[i+1:]...)
			if exchange.autodelete && len(exchange.bindings) == 0 {
				go exchange.autodeleteTimeout()
			}
			return nil
		}
	}
	return nil
}

func (exchange *Exchange) autodeleteTimeout() {
	// There's technically a race condition here where a new binding could be
	// added right as we check this, but after a 5 second wait with no activity
	// I think this is probably safe enough.
	var now = time.Now()
	exchange.deleteActive = now
	time.Sleep(5 * time.Second)
	if exchange.deleteActive == now {
		exchange.deleteChan <- exchange
	}
}
