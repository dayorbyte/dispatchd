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

func (exchange *Exchange) start() {
	go func() {
		for {
			<-exchange.incoming
			fmt.Printf("Exchange %s got a frame\n", exchange.name)
		}
	}()
}

func (exchange *Exchange) publish(server *Server, channel *Channel, msg *Message) {
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
		channel.sendContent(&amqp.BasicReturn{
			ReplyCode: 200, // TODO: what code?
			ReplyText: "Message unroutable",
		}, msg)
	}

}

func (exchange *Exchange) addBinding(method *amqp.QueueBind, fromDisk bool) (uint16, error) {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	// Check queue
	var queue, foundQueue = exchange.server.queues[method.Queue]
	if !foundQueue {
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

func (exchange *Exchange) removeBinding(queue *Queue, binding *Binding) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	err := exchange.server.depersistBinding(&amqp.QueueBind{
		Exchange:   exchange.name,
		Queue:      queue.name,
		RoutingKey: binding.key,
		Arguments:  binding.arguments,
	})
	if err != nil {
		return err
	}
	for i, b := range exchange.bindings {
		if binding.Equals(b) {
			exchange.bindings = append(exchange.bindings[:i], exchange.bindings[i+1:]...)
			return nil
		}
	}
	return nil
}
