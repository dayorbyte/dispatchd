package main

import (
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) queueRoute(methodFrame amqp.MethodFrame) *AMQPError {
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
	return NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) queueDeclare(method *amqp.QueueDeclare) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	// No name means generate a name
	if len(method.Queue) == 0 {
		method.Queue = randomId()
	}

	// Check the name format
	var err = amqp.CheckExchangeOrQueueName(method.Queue)
	if err != nil {
		return NewSoftError(406, err.Error(), classId, methodId)
	}

	if method.Passive {
		queue, found := channel.conn.server.queues[method.Queue]
		if found {
			if !method.NoWait {
				var qsize = uint32(queue.queue.Len())
				var csize = queue.activeConsumerCount()
				channel.SendMethod(&amqp.QueueDeclareOk{method.Queue, qsize, csize})
			}
			channel.lastQueueName = method.Queue
			return nil
		}
		return NewSoftError(404, "Queue not found", classId, methodId)
	}

	name, err := channel.conn.server.declareQueue(method, channel.conn.id, false)
	if err != nil {
		return NewSoftError(500, "Error creating queue", classId, methodId)
	}
	channel.lastQueueName = method.Queue
	if !method.NoWait {
		channel.SendMethod(&amqp.QueueDeclareOk{name, uint32(0), uint32(0)})
	}
	return nil
}

func (channel *Channel) queueBind(method *amqp.QueueBind) *AMQPError {
	var classId, methodId = method.MethodIdentifier()

	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	// Check exchange
	var exchange, foundExchange = channel.server.exchanges[method.Exchange]
	if !foundExchange {
		return NewSoftError(404, "Exchange not found", classId, methodId)
	}

	// Check queue
	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue || queue.closed {
		return NewSoftError(404, fmt.Sprintf("Queue not found: %s", method.Queue), classId, methodId)
	}

	if queue.connId != -1 && queue.connId != channel.conn.id {
		return NewSoftError(405, fmt.Sprintf("Queue is locked to another connection"), classId, methodId)
	}

	// Add binding
	err := exchange.addBinding(method, channel.conn.id, false)
	if err != nil {
		return NewSoftError(500, err.Error(), classId, methodId)
	}

	// Persist durable bindings
	if exchange.durable && queue.durable {
		var err = channel.server.persistBinding(method)
		if err != nil {
			return NewSoftError(500, err.Error(), classId, methodId)
		}
	}

	if !method.NoWait {
		channel.SendMethod(&amqp.QueueBindOk{})
	}
	return nil
}

func (channel *Channel) queuePurge(method *amqp.QueuePurge) *AMQPError {
	fmt.Println("Got queuePurge")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		return NewSoftError(404, "Queue not found", classId, methodId)
	}

	if queue.connId != -1 && queue.connId != channel.conn.id {
		return NewSoftError(405, "Queue is locked to another connection", classId, methodId)
	}

	numPurged := queue.purge()
	if !method.NoWait {
		channel.SendMethod(&amqp.QueuePurgeOk{numPurged})
	}
	return nil
}

func (channel *Channel) queueDelete(method *amqp.QueueDelete) *AMQPError {
	fmt.Println("Got queueDelete")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	numPurged, errCode, err := channel.server.deleteQueue(method, channel.conn.id)
	if err != nil {
		return NewSoftError(errCode, err.Error(), classId, methodId)
	}

	if !method.NoWait {
		channel.SendMethod(&amqp.QueueDeleteOk{numPurged})
	}
	return nil
}

func (channel *Channel) queueUnbind(method *amqp.QueueUnbind) *AMQPError {
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		return NewSoftError(404, "Queue not found", classId, methodId)
	}

	if queue.connId != -1 && queue.connId != channel.conn.id {
		return NewSoftError(405, "Queue is locked to another connection", classId, methodId)
	}

	// Check exchange
	var exchange, foundExchange = channel.server.exchanges[method.Exchange]
	if !foundExchange {
		return NewSoftError(404, "Exchange not found", classId, methodId)
	}

	var binding = NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments)

	if err := exchange.removeBinding(channel.server, queue, binding); err != nil {
		return NewSoftError(500, err.Error(), classId, methodId)
	}
	channel.SendMethod(&amqp.QueueUnbindOk{})
	return nil
}
