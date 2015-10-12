package main

import (
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
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

func exchangeNameToType(et string) (extype, error) {
	switch {
	case et == "direct":
		return EX_TYPE_DIRECT, nil
	case et == "fanout":
		return EX_TYPE_FANOUT, nil
	default:
		panic("bad exchange type! " + et)
		return 0, errors.New("Unknown exchang type " + et)
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

func (exchange *Exchange) publish(server *Server, method *amqp.BasicPublish, header *amqp.ContentHeaderFrame, bodyFrames []*amqp.WireFrame) {
	fmt.Printf("Got message in exchange %s\n", exchange.name)
	switch {
	case exchange.extype == EX_TYPE_DIRECT:
		for _, binding := range exchange.bindings {
			if binding.matchDirect(method) {
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(&Message{
					header:   header,
					payload:  bodyFrames,
					exchange: method.Exchange,
					key:      method.RoutingKey,
				})
			}
		}
	case exchange.extype == EX_TYPE_FANOUT:
		for _, binding := range exchange.bindings {
			if binding.matchFanout(method) {
				var queue, foundQueue = server.queues[binding.queueName]
				if !foundQueue {
					panic("queue not found!")
				}
				queue.add(&Message{
					header:   header,
					payload:  bodyFrames,
					exchange: method.Exchange,
					key:      method.RoutingKey,
				})
			}
		}
	default:
		panic("unknown exchange type!")
	}

}

func (exchange *Exchange) addBinding(queue *Queue, binding *Binding) bool {
	for _, b := range exchange.bindings {
		if binding == b {
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
