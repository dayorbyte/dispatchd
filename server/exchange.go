package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"reflect"
)

type extype uint8

const (
	EX_TYPE_DIRECT extype = 1
	EX_TYPE_FANOUT extype = 2
	EX_TYPE_TOPIC  extype = 3
)

type Exchange struct {
	name       string
	extype     extype
	durable    bool
	autodelete bool
	internal   bool
	arguments  amqp.Table
	incoming   chan amqp.Frame
	system     bool
	bindings   []*Binding
}

func (exchange *Exchange) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": exchangeTypeToName(exchange.extype),
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
	if reflect.DeepEqual(ex1, ex2) {
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
	default:
		return 0, errors.New("Unknown exchang type " + et)
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
	var matched = false
	switch {
	case exchange.extype == EX_TYPE_DIRECT:
		for _, binding := range exchange.bindings {
			if binding.matchDirect(msg.method) {
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(channel, msg)
				matched = true
			}
		}
	case exchange.extype == EX_TYPE_FANOUT:
		for _, binding := range exchange.bindings {
			if binding.matchFanout(msg.method) {
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(channel, msg)
				matched = true
			}
		}
	case exchange.extype == EX_TYPE_TOPIC:
		for _, binding := range exchange.bindings {
			if binding.matchTopic(msg.method) {
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(channel, msg)
				matched = true
			}
		}
	default:
		panic("unknown exchange type!")
	}
	if matched == true {
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

func (exchange *Exchange) addBinding(queue *Queue, binding *Binding) bool {
	for _, b := range exchange.bindings {
		if binding.Equals(b) {
			return false
		}
	}
	exchange.bindings = append(exchange.bindings, binding)
	return true
}

func (exchange *Exchange) removeBinding(queue *Queue, binding *Binding) bool {
	for i, b := range exchange.bindings {
		if binding == b {
			exchange.bindings = append(exchange.bindings[:i], exchange.bindings[i+1:]...)
			return true
		}
	}
	return false
}
