package server

import (
	"fmt"
	"github.com/jeffjenkins/dispatchd/amqp"
	"github.com/jeffjenkins/dispatchd/binding"
	"github.com/jeffjenkins/dispatchd/queue"
	"github.com/jeffjenkins/dispatchd/util"
)

func (channel *Channel) queueRoute(methodFrame amqp.MethodFrame) *amqp.AMQPError {
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
	return amqp.NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) queueDeclare(method *amqp.QueueDeclare) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	// No name means generate a name
	if len(method.Queue) == 0 {
		method.Queue = util.RandomId()
	}

	// Check the name format
	var err = amqp.CheckExchangeOrQueueName(method.Queue)
	if err != nil {
		return amqp.NewSoftError(406, err.Error(), classId, methodId)
	}

	// If this is a passive request, do the appropriate checks and return
	if method.Passive {
		queue, found := channel.conn.server.queues[method.Queue]
		if found {
			if !method.NoWait {
				var qsize = uint32(queue.Len())
				var csize = queue.ActiveConsumerCount()
				channel.SendMethod(&amqp.QueueDeclareOk{method.Queue, qsize, csize})
			}
			channel.lastQueueName = method.Queue
			return nil
		}
		return amqp.NewSoftError(404, "Queue not found", classId, methodId)
	}

	// Create the new queue
	var connId = channel.conn.id
	if !method.Exclusive {
		connId = -1
	}
	var queue = queue.NewQueue(
		method.Queue,
		method.Durable,
		method.Exclusive,
		method.AutoDelete,
		method.Arguments,
		connId,
		channel.server.msgStore,
		channel.server.queueDeleter,
	)

	// If the new queue exists already, ensure the settings are the same. If it
	// doesn't, add it and optionally persist it
	existing, hasKey := channel.server.queues[queue.Name]
	if hasKey {
		if existing.ConnId != -1 && existing.ConnId != channel.conn.id {
			return amqp.NewSoftError(405, "Queue is locked to another connection", classId, methodId)
		}
		if !existing.EquivalentQueues(queue) {
			return amqp.NewSoftError(406, "Queue exists and is not equivalent to existing", classId, methodId)
		}
	} else {
		err = channel.server.addQueue(queue)
		if err != nil { // pragma: nocover
			return amqp.NewSoftError(500, "Error creating queue", classId, methodId)
		}
		// Persist
		if queue.Durable {
			queue.Persist(channel.server.db)
		}
	}

	channel.lastQueueName = method.Queue
	if !method.NoWait {
		channel.SendMethod(&amqp.QueueDeclareOk{queue.Name, uint32(0), uint32(0)})
	}
	return nil
}

func (channel *Channel) queueBind(method *amqp.QueueBind) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()

	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return amqp.NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	// Check exchange
	var exchange, foundExchange = channel.server.exchanges[method.Exchange]
	if !foundExchange {
		return amqp.NewSoftError(404, "Exchange not found", classId, methodId)
	}

	// Check queue
	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue || queue.Closed {
		return amqp.NewSoftError(404, fmt.Sprintf("Queue not found: %s", method.Queue), classId, methodId)
	}

	if queue.ConnId != -1 && queue.ConnId != channel.conn.id {
		return amqp.NewSoftError(405, fmt.Sprintf("Queue is locked to another connection"), classId, methodId)
	}

	// Create binding
	b, err := binding.NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments, exchange.IsTopic())
	if err != nil {
		return amqp.NewSoftError(500, err.Error(), classId, methodId)
	}

	// Add binding
	err = exchange.AddBinding(b, channel.conn.id)
	if err != nil {
		return amqp.NewSoftError(500, err.Error(), classId, methodId)
	}

	// Persist durable bindings
	if exchange.Durable && queue.Durable {
		var err = b.Persist(channel.server.db)
		if err != nil {
			return amqp.NewSoftError(500, err.Error(), classId, methodId)
		}
	}

	if !method.NoWait {
		channel.SendMethod(&amqp.QueueBindOk{})
	}
	return nil
}

func (channel *Channel) queuePurge(method *amqp.QueuePurge) *amqp.AMQPError {
	fmt.Println("Got queuePurge")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return amqp.NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		return amqp.NewSoftError(404, "Queue not found", classId, methodId)
	}

	if queue.ConnId != -1 && queue.ConnId != channel.conn.id {
		return amqp.NewSoftError(405, "Queue is locked to another connection", classId, methodId)
	}

	numPurged := queue.Purge()
	if !method.NoWait {
		channel.SendMethod(&amqp.QueuePurgeOk{numPurged})
	}
	return nil
}

func (channel *Channel) queueDelete(method *amqp.QueueDelete) *amqp.AMQPError {
	fmt.Println("Got queueDelete")
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return amqp.NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	numPurged, errCode, err := channel.server.deleteQueue(method, channel.conn.id)
	if err != nil {
		return amqp.NewSoftError(errCode, err.Error(), classId, methodId)
	}

	if !method.NoWait {
		channel.SendMethod(&amqp.QueueDeleteOk{numPurged})
	}
	return nil
}

func (channel *Channel) queueUnbind(method *amqp.QueueUnbind) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()

	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return amqp.NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}

	var queue, foundQueue = channel.server.queues[method.Queue]
	if !foundQueue {
		return amqp.NewSoftError(404, "Queue not found", classId, methodId)
	}

	if queue.ConnId != -1 && queue.ConnId != channel.conn.id {
		return amqp.NewSoftError(405, "Queue is locked to another connection", classId, methodId)
	}

	// Check exchange
	var exchange, foundExchange = channel.server.exchanges[method.Exchange]
	if !foundExchange {
		return amqp.NewSoftError(404, "Exchange not found", classId, methodId)
	}

	var binding, err = binding.NewBinding(
		method.Queue,
		method.Exchange,
		method.RoutingKey,
		method.Arguments,
		exchange.IsTopic(),
	)

	if err != nil {
		return amqp.NewSoftError(500, err.Error(), classId, methodId)
	}

	if queue.Durable && exchange.Durable {
		err := binding.Depersist(channel.server.db)
		if err != nil {
			return amqp.NewSoftError(500, "Could not de-persist binding!", classId, methodId)
		}
	}

	if err := exchange.RemoveBinding(binding); err != nil {
		return amqp.NewSoftError(500, err.Error(), classId, methodId)
	}
	channel.SendMethod(&amqp.QueueUnbindOk{})
	return nil
}
