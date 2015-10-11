package main

import (
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) queueRoute(methodFrame amqp.MethodFrame) error {
	switch method := methodFrame.(type) {
	case *amqp.QueueDeclare:
		return channel.queueDeclare(method)
	case *amqp.QueueBind:
		return channel.queueBind(method)
	case *amqp.QueuePurge:
		return channel.queuePurge(method)
	case *amqp.QueueDelete:
		return channel.queueDelete(method)
	case *amqp.QueueUnbind:
		return channel.queueUnbind(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) queueDeclare(method *amqp.QueueDeclare) error {
	fmt.Println("Got queueDeclare")
	if method.Passive {
		_, found := channel.conn.server.queues[method.Queue]
		if found {
			channel.sendMethod(&amqp.QueueDeclareOk{method.Queue, 0, 0})
			return nil
		}
		var classId, methodId = method.MethodIdentifier()
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
	}
	fmt.Println("calling declareQueue")
	channel.conn.server.declareQueue(method)
	fmt.Println("Sending QueueDeclareOk")
	if !method.NoWait {
		channel.sendMethod(&amqp.QueueDeclareOk{method.Queue, uint32(0), uint32(0)})
	}
	return nil
}

func (channel *Channel) queueBind(method *amqp.QueueBind) error {
	fmt.Println("Got queueBind")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
		return nil
	}
	// Check exchange
	var exchange, foundExchange = channel.server.exchanges[method.Exchange]
	if !foundExchange {
		channel.channelErrorWithMethod(404, "Exchange not found", classId, methodId)
		return nil
	}

	var binding = &Binding{
		queueName:    method.Queue,
		exchangeName: method.Exchange,
		key:          method.RoutingKey,
		arguments:    method.Arguments,
	}

	exchange.addBinding(queue, binding)

	if !method.NoWait {
		channel.sendMethod(&amqp.QueueBindOk{})
	}
	return nil
}

func (channel *Channel) queuePurge(method *amqp.QueuePurge) error {
	fmt.Println("Got queuePurge")
	var classId, methodId = method.MethodIdentifier()

	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
		return nil
	}
	numPurged := queue.purge()
	if !method.NoWait {
		channel.sendMethod(&amqp.QueuePurgeOk{numPurged})
	}
	return nil
}

func (channel *Channel) queueDelete(method *amqp.QueueDelete) error {
	fmt.Println("Got queueDelete")
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.sendMethod(&amqp.QeueDeleteOk{0}) // TODO(MUST): num messages deleted
	// }
	return nil
}

func (channel *Channel) queueUnbind(method *amqp.QueueUnbind) error {
	fmt.Println("Got queueUnbind")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
		return nil
	}
	// Check exchange
	var exchange, foundExchange = channel.server.exchanges[method.Exchange]
	if !foundExchange {
		channel.channelErrorWithMethod(404, "Exchange not found", classId, methodId)
		return nil
	}

	var binding = &Binding{
		queueName:    method.Queue,
		exchangeName: method.Exchange,
		key:          method.RoutingKey,
		arguments:    method.Arguments,
	}

	exchange.removeBinding(queue, binding)
	return nil
}
