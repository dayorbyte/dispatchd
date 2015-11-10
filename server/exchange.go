package main

import (
	"encoding/json"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
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
	server       *Server
	closed       bool
	deleteActive time.Time
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

func (exchange *Exchange) queuesForPublish(server *Server, channel *Channel, msg *amqp.Message) map[string]*Queue {
	var queues = make(map[string]*Queue)
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

func (exchange *Exchange) publish(server *Server, channel *Channel, msg *amqp.Message) {
	// Concurrency note: Since there is no lock we can, technically, have messages
	// published after the exchange has been closed. These couldn't be on the same
	// channel as the close is happening on, so that seems justifiable.
	if exchange.closed {
		if msg.Method.Mandatory || msg.Method.Immediate {
			exchange.returnMessage(channel, msg, 313, "Exchange closed, cannot route to queues or consumers")
		}
		return
	}
	queues := exchange.queuesForPublish(server, channel, msg)

	if len(queues) == 0 {
		// If we got here the message was unroutable.
		if msg.Method.Mandatory || msg.Method.Immediate {
			exchange.returnMessage(channel, msg, 313, "No queues available")
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
		queueMessagesByQueue, err := server.msgStore.AddMessage(msg, queueNames)
		if err != nil {
			channel.channelErrorWithMethod(500, err.Error(), 60, 40)
			return
		}
		// Try to immediately consumed it
		for name, queue := range queues {
			qms := queueMessagesByQueue[name]
			for _, qm := range qms {
				var oneConsumed = queue.consumeImmediate(qm)
				if !oneConsumed {
					server.msgStore.RemoveRef(qm.Id, name)
				}
				consumed = oneConsumed || consumed
			}
		}
		if !consumed {
			exchange.returnMessage(channel, msg, 313, "No consumers available for immediate message")
		}
		return
	}

	// Add the message to the message store along with the queues we're about to add it to
	queueMessagesByQueue, err := server.msgStore.AddMessage(msg, queueNames)
	if err != nil {
		channel.channelErrorWithMethod(500, err.Error(), 60, 40)
		return
	}

	for name, queue := range queues {
		qms := queueMessagesByQueue[name]
		for _, qm := range qms {
			if !queue.add(qm) {
				// If we couldn't add it means the queue is closed and we should
				// remove the ref from the message store. The queue being closed means
				// it is going away, so worst case if the server dies we have to process
				// and discard the message on boot.
				server.msgStore.RemoveRef(msg.Id, name)
			}
		}
	}
}

func (exchange *Exchange) returnMessage(channel *Channel, msg *amqp.Message, code uint16, text string) {
	channel.sendContent(&amqp.BasicReturn{
		Exchange:   exchange.name,
		RoutingKey: msg.Method.RoutingKey,
		ReplyCode:  code,
		ReplyText:  text,
	}, msg)
}

func (exchange *Exchange) addBinding(method *amqp.QueueBind, connId int64, fromDisk bool) (uint16, error) {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	// Check queue
	var queue, foundQueue = exchange.server.queues[method.Queue]
	if !foundQueue || queue.closed {
		return 404, fmt.Errorf("Queue not found: %s", method.Queue)
	}

	if queue.connId != -1 && queue.connId != connId {
		return 405, fmt.Errorf("Queue is locked to another connection")
	}

	var binding = NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments)

	for _, b := range exchange.bindings {
		if binding.Equals(b) {
			return 0, nil
		}
	}
	if exchange.durable && queue.durable && !fromDisk {
		var err = exchange.server.persistBinding(method)
		if err != nil {
			return 500, err
		}
	}
	if exchange.autodelete {
		exchange.deleteActive = time.Unix(0, 0)
	}
	exchange.bindings = append(exchange.bindings, binding)
	return 0, nil
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

func (exchange *Exchange) removeBinding(server *Server, queue *Queue, binding *Binding) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	// First de-persist
	if queue.durable && exchange.durable {
		err := exchange.server.depersistBinding(binding)
		if err != nil {
			return err
		}
	}

	// Delete binding
	for i, b := range exchange.bindings {
		if binding.Equals(b) {
			exchange.bindings = append(exchange.bindings[:i], exchange.bindings[i+1:]...)
			if exchange.autodelete && len(exchange.bindings) == 0 {
				go exchange.autodeleteTimeout(server)
			}
			return nil
		}
	}
	return nil
}

func (exchange *Exchange) autodeleteTimeout(server *Server) {
	// There's technically a race condition here where a new binding could be
	// added right as we check this, but after a 5 second wait with no activity
	// I think this is probably safe enough.
	var now = time.Now()
	exchange.deleteActive = now
	time.Sleep(5 * time.Second)
	if exchange.deleteActive == now {
		server.deleteExchange(&amqp.ExchangeDelete{
			Exchange: exchange.name,
		})
	}
}
