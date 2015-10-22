package main

import (
	"fmt"
	"math/rand"
	"time"

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
	var classId, methodId = method.MethodIdentifier()
	// No name means generate a name
	if len(method.Queue) == 0 {
		method.Queue = randomQueueId()
	}

	// Check the name format
	var err = amqp.CheckExchangeOrQueueName(method.Queue)
	if err != nil {
		channel.channelErrorWithMethod(406, err.Error(), classId, methodId)
		return nil
	}

	if method.Passive {
		queue, found := channel.conn.server.queues[method.Queue]
		if found {
			if !method.NoWait {
				var qsize = uint32(queue.queue.Len())
				var csize = queue.activeConsumerCount()
				channel.sendMethod(&amqp.QueueDeclareOk{method.Queue, qsize, csize})
			}
			channel.lastQueueName = method.Queue
			return nil
		}
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
	}
	name, err := channel.conn.server.declareQueue(method)
	if err != nil {
		channel.channelErrorWithMethod(500, "Error creating queue", classId, methodId)
		return nil
	}
	channel.lastQueueName = method.Queue
	if !method.NoWait {
		channel.sendMethod(&amqp.QueueDeclareOk{name, uint32(0), uint32(0)})
	}
	return nil
}

func (channel *Channel) queueBind(method *amqp.QueueBind) error {
	var classId, methodId = method.MethodIdentifier()

	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
			return nil
		} else {
			method.Queue = channel.lastQueueName
		}
	}

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

	var binding = NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments)

	exchange.addBinding(queue, binding)

	if !method.NoWait {
		channel.sendMethod(&amqp.QueueBindOk{})
	}
	return nil
}

func (channel *Channel) queuePurge(method *amqp.QueuePurge) error {
	fmt.Println("Got queuePurge")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
			return nil
		} else {
			method.Queue = channel.lastQueueName
		}
	}

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

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
			return nil
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	numPurged, errCode, err := channel.server.deleteQueue(method)
	if err != nil {
		channel.channelErrorWithMethod(errCode, err.Error(), classId, methodId)
		return nil
	}

	if !method.NoWait {
		channel.sendMethod(&amqp.QueueDeleteOk{numPurged})
	}
	return nil
}

func (channel *Channel) queueUnbind(method *amqp.QueueUnbind) error {
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
			return nil
		} else {
			method.Queue = channel.lastQueueName
		}
	}

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

	var binding = NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments)

	exchange.removeBinding(queue, binding)
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randomQueueId() string {
	var size = 32
	var numChars = len(chars)
	id := make([]rune, size)
	for i := range id {
		id[i] = chars[rand.Intn(numChars)]
	}
	return string(id)
}
