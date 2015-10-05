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
		channel.conn.connectionErrorWithMethod(404, "Not found", classId, methodId)
	}
	fmt.Println("calling declareQueue")
	channel.conn.server.declareQueue(method)
	fmt.Println("Sending QueueDeclareOk")
	channel.sendMethod(&amqp.QueueDeclareOk{method.Queue, uint32(0), uint32(0)})
	return nil
}

func (channel *Channel) queueBind(method *amqp.QueueBind) error {
	fmt.Println("Got queueBind")
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) queuePurge(method *amqp.QueuePurge) error {
	fmt.Println("Got queuePurge")
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) queueDelete(method *amqp.QueueDelete) error {
	fmt.Println("Got queueDelete")
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) queueUnbind(method *amqp.QueueUnbind) error {
	fmt.Println("Got queueUnbind")
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}
