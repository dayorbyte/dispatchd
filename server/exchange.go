package main

import (
	"encoding/json"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
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
	arguments    amqp.Table
	system       bool
	bindings     []*Binding
	bindingsLock sync.Mutex
	incoming     chan amqp.Frame
	server       *Server
	closed       bool
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
	if ex1.name != ex2.name {
		return false
	}
	if ex1.extype != ex2.extype {
		return false
	}
	if ex1.durable != ex2.durable {
		return false
	}
	if ex1.autodelete != ex2.autodelete {
		return false
	}
	if ex1.internal != ex2.internal {
		return false
	}
	if !amqp.EquivalentTables(&ex1.arguments, &ex2.arguments) {
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

func (exchange *Exchange) publish(server *Server, channel *Channel, msg *Message) {
	// Concurrency note: Since there is no lock we can, technically, have messages
	// published after the exchange has been closed. These couldn't be on the same
	// channel as the close is happening on, so that seems justifiable.
	if exchange.closed {
		if msg.method.Mandatory || msg.method.Immediate {
			returnMessage(channel, msg)
		}
		return
	}
	var seen = make(map[string]bool)
	switch {
	case exchange.extype == EX_TYPE_DIRECT:
		for _, binding := range exchange.bindings {
			if binding.matchDirect(msg.method) {
				var _, alreadySeen = seen[binding.queueName]
				if alreadySeen {
					continue
				}
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(channel, msg)
				seen[binding.queueName] = true
			}
		}
	case exchange.extype == EX_TYPE_FANOUT:
		for _, binding := range exchange.bindings {
			if binding.matchFanout(msg.method) {
				var _, alreadySeen = seen[binding.queueName]
				if alreadySeen {
					continue
				}
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(channel, msg)
				seen[binding.queueName] = true
			}
		}
	case exchange.extype == EX_TYPE_TOPIC:
		for _, binding := range exchange.bindings {
			if binding.matchTopic(msg.method) {
				var _, alreadySeen = seen[binding.queueName]
				if alreadySeen {
					continue
				}
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(channel, msg)
				seen[binding.queueName] = true
			}
		}
	case exchange.extype == EX_TYPE_HEADERS:
		// TODO: implement
		panic("Headers is not implemented!")
	default:
		// TODO: can this happen? Seems like checks should be earlier
		panic("unknown exchange type!")
	}
	if len(seen) > 0 {
		return
	}
	// If we got here the message was unroutable.
	if msg.method.Mandatory || msg.method.Immediate {
		returnMessage(channel, msg)
	}
}

func returnMessage(channel *Channel, msg *Message) {
	channel.sendContent(&amqp.BasicReturn{
		ReplyCode: 200, // TODO: what code?
		ReplyText: "Message unroutable",
	}, msg)
}

func (exchange *Exchange) addBinding(method *amqp.QueueBind, fromDisk bool) (uint16, error) {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	// Check queue
	var queue, foundQueue = exchange.server.queues[method.Queue]
	if !foundQueue || queue.closed {
		return 404, fmt.Errorf("Queue not found: %s", method.Queue)
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

func (exchange *Exchange) removeBinding(queue *Queue, binding *Binding) error {
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
			return nil
		}
	}
	return nil
}
