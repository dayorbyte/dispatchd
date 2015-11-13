package main

import (
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/interfaces"
)

func (channel *Channel) basicRoute(methodFrame amqp.MethodFrame) error {
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
	return errors.New("Unable to route method frame")
}

func (channel *Channel) basicQos(method *amqp.BasicQos) error {
	channel.setPrefetch(method.PrefetchCount, method.PrefetchSize, method.Global)
	channel.sendMethod(&amqp.BasicQosOk{})
	return nil
}

func (channel *Channel) basicRecover(method *amqp.BasicRecover) error {
	channel.recover(method.Requeue)
	channel.sendMethod(&amqp.BasicRecoverOk{})
	return nil
}

func (channel *Channel) basicNack(method *amqp.BasicNack) error {
	var ok = false
	if method.Multiple {
		ok = channel.nackBelow(method.DeliveryTag, method.Requeue, false)
	} else {
		ok = channel.nackOne(method.DeliveryTag, method.Requeue, false)
	}
	if !ok {
		var classId, methodId = method.MethodIdentifier()
		var msg = fmt.Sprintf("Precondition Failed: Delivery Tag not found: %d", method.DeliveryTag)
		channel.channelErrorWithMethod(406, msg, classId, methodId)
	}
	return nil
}

func (channel *Channel) basicConsume(method *amqp.BasicConsume) error {
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
	// TODO: do not directly access channel.conn.server.queues
	var queue, found = channel.conn.server.queues[method.Queue]
	if !found {
		// Spec doesn't say, but seems like a 404?
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
	}
	if len(method.ConsumerTag) == 0 {
		method.ConsumerTag = randomId()
	}
	errCode, err := queue.addConsumer(channel, method)
	if err != nil {
		var classId, methodId = method.MethodIdentifier()
		// TODO: there should probably be a single error handling function
		// which does channel or connection errors based on the code. I would
		// need to gen a hard vs soft error function, but it wouldn't be too
		// difficult
		channel.conn.connectionErrorWithMethod(errCode, err.Error(), classId, methodId)
	}
	if !method.NoWait {
		channel.sendMethod(&amqp.BasicConsumeOk{method.ConsumerTag})
	}

	return nil
}

func (channel *Channel) basicCancel(method *amqp.BasicCancel) error {

	if err := channel.removeConsumer(method.ConsumerTag); err != nil {
		var classId, methodId = method.MethodIdentifier()
		channel.channelErrorWithMethod(404, "Consumer not found", classId, methodId)
		return nil
	}

	if !method.NoWait {
		channel.sendMethod(&amqp.BasicCancelOk{method.ConsumerTag})
	}
	return nil
}

func (channel *Channel) basicCancelOk(method *amqp.BasicCancelOk) error {
	// TODO(MAY)
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	return nil
}

func (channel *Channel) basicPublish(method *amqp.BasicPublish) error {
	var _, found = channel.server.exchanges[method.Exchange]
	if !found {
		var classId, methodId = method.MethodIdentifier()
		channel.channelErrorWithMethod(404, "Exchange not found", classId, methodId)
		return nil
	}
	channel.startPublish(method)
	return nil
}

func (channel *Channel) basicGet(method *amqp.BasicGet) error {
	// var classId, methodId = method.MethodIdentifier()
	// channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	var queue, found = channel.conn.server.queues[method.Queue]
	if !found {
		// Spec doesn't say, but seems like a 404?
		var classId, methodId = method.MethodIdentifier()
		channel.channelErrorWithMethod(404, "Queue not found", classId, methodId)
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

func (channel *Channel) basicAck(method *amqp.BasicAck) error {
	var ok = false
	if method.Multiple {
		ok = channel.ackBelow(method.DeliveryTag, false)
	} else {
		ok = channel.ackOne(method.DeliveryTag, false)
	}
	if !ok {
		var classId, methodId = method.MethodIdentifier()
		var msg = fmt.Sprintf("Precondition Failed: Delivery Tag not found: %d", method.DeliveryTag)
		channel.channelErrorWithMethod(406, msg, classId, methodId)
	}
	return nil
}

func (channel *Channel) basicReject(method *amqp.BasicReject) error {
	if !channel.nackOne(method.DeliveryTag, method.Requeue, false) {
		var classId, methodId = method.MethodIdentifier()
		var msg = fmt.Sprintf("Precondition Failed: Delivery Tag not found: %d", method.DeliveryTag)
		channel.channelErrorWithMethod(406, msg, classId, methodId)
	}
	return nil
}
