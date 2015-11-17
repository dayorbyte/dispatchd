package main

import (
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/interfaces"
	"github.com/jeffjenkins/mq/stats"
)

func (channel *Channel) basicRoute(methodFrame amqp.MethodFrame) *AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.BasicQos:
		return channel.basicQos(method)
	case *amqp.BasicRecover:
		return channel.basicRecover(method)
	case *amqp.BasicNack:
		return channel.basicNack(method)
	case *amqp.BasicConsume:
		return channel.basicConsume(method)
	case *amqp.BasicCancel:
		return channel.basicCancel(method)
	case *amqp.BasicCancelOk:
		return channel.basicCancelOk(method)
	case *amqp.BasicPublish:
		return channel.basicPublish(method)
	case *amqp.BasicGet:
		return channel.basicGet(method)
	case *amqp.BasicAck:
		return channel.basicAck(method)
	case *amqp.BasicReject:
		return channel.basicReject(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) basicQos(method *amqp.BasicQos) *AMQPError {
	channel.setPrefetch(method.PrefetchCount, method.PrefetchSize, method.Global)
	channel.sendMethod(&amqp.BasicQosOk{})
	return nil
}

func (channel *Channel) basicRecover(method *amqp.BasicRecover) *AMQPError {
	channel.recover(method.Requeue)
	channel.sendMethod(&amqp.BasicRecoverOk{})
	return nil
}

func (channel *Channel) basicNack(method *amqp.BasicNack) *AMQPError {
	if method.Multiple {
		return channel.nackBelow(method.DeliveryTag, method.Requeue, false)
	}
	return channel.nackOne(method.DeliveryTag, method.Requeue, false)
}

func (channel *Channel) basicConsume(method *amqp.BasicConsume) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	// Check queue
	if len(method.Queue) == 0 {
		if len(channel.lastQueueName) == 0 {
			return NewSoftError(404, "Queue not found", classId, methodId)
		} else {
			method.Queue = channel.lastQueueName
		}
	}
	// TODO: do not directly access channel.conn.server.queues
	var queue, found = channel.conn.server.queues[method.Queue]
	if !found {
		// Spec doesn't say, but seems like a 404?
		return NewSoftError(404, "Queue not found", classId, methodId)
	}
	if len(method.ConsumerTag) == 0 {
		method.ConsumerTag = randomId()
	}
	errCode, err := queue.addConsumer(channel, method)
	if err != nil {
		var classId, methodId = method.MethodIdentifier()
		return NewHardError(errCode, err.Error(), classId, methodId)
	}
	if !method.NoWait {
		channel.sendMethod(&amqp.BasicConsumeOk{method.ConsumerTag})
	}

	return nil
}

func (channel *Channel) basicCancel(method *amqp.BasicCancel) *AMQPError {

	if err := channel.removeConsumer(method.ConsumerTag); err != nil {
		var classId, methodId = method.MethodIdentifier()
		return NewSoftError(404, "Consumer not found", classId, methodId)
	}

	if !method.NoWait {
		channel.sendMethod(&amqp.BasicCancelOk{method.ConsumerTag})
	}
	return nil
}

func (channel *Channel) basicCancelOk(method *amqp.BasicCancelOk) *AMQPError {
	// TODO(MAY)
	var classId, methodId = method.MethodIdentifier()
	return NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) basicPublish(method *amqp.BasicPublish) *AMQPError {
	defer stats.RecordHisto(channel.statPublish, stats.Start())
	var _, found = channel.server.exchanges[method.Exchange]
	if !found {
		var classId, methodId = method.MethodIdentifier()
		return NewSoftError(404, "Exchange not found", classId, methodId)
	}
	channel.startPublish(method)
	return nil
}

func (channel *Channel) basicGet(method *amqp.BasicGet) *AMQPError {
	// var classId, methodId = method.MethodIdentifier()
	// channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	var queue, found = channel.conn.server.queues[method.Queue]
	if !found {
		// Spec doesn't say, but seems like a 404?
		var classId, methodId = method.MethodIdentifier()
		return NewSoftError(404, "Queue not found", classId, methodId)
	}
	var qm = queue.getOneForced()
	if qm == nil {
		channel.sendMethod(&amqp.BasicGetEmpty{})
		return nil
	}

	var rhs = []interfaces.MessageResourceHolder{channel}
	msg, err := channel.server.msgStore.GetAndDecrRef(qm, queue.name, rhs)
	if err != nil {
		// TODO: return 500 error
		channel.sendMethod(&amqp.BasicGetEmpty{})
		return nil
	}

	channel.sendContent(&amqp.BasicGetOk{
		DeliveryTag:  channel.nextDeliveryTag(),
		Redelivered:  qm.DeliveryCount > 0,
		Exchange:     msg.Exchange,
		RoutingKey:   msg.Key,
		MessageCount: 1,
	}, msg)
	return nil
}

func (channel *Channel) basicAck(method *amqp.BasicAck) *AMQPError {
	if method.Multiple {
		return channel.ackBelow(method.DeliveryTag, false)
	}
	return channel.ackOne(method.DeliveryTag, false)
}

func (channel *Channel) basicReject(method *amqp.BasicReject) *AMQPError {
	return channel.nackOne(method.DeliveryTag, method.Requeue, false)
}
