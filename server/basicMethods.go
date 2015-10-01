package main

import (
	"errors"
	"fmt"
	// "bytes"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) basicRoute(methodFrame amqp.MethodFrame) error {
	switch method := methodFrame.(type) {
	case *amqp.BasicQos:
		return channel.basicQos(method)
	case *amqp.BasicRecoverAsync:
		return channel.basicRecoverAsync(method)
	case *amqp.BasicQosOk:
		return channel.basicQosOk(method)
	case *amqp.BasicRecover:
		return channel.basicRecover(method)
	case *amqp.BasicRecoverOk:
		return channel.basicRecoverOk(method)
	case *amqp.BasicNack:
		return channel.basicNack(method)
	case *amqp.BasicConsume:
		return channel.basicConsume(method)
	case *amqp.BasicConsumeOk:
		return channel.basicConsumeOk(method)
	case *amqp.BasicCancel:
		return channel.basicCancel(method)
	case *amqp.BasicCancelOk:
		return channel.basicCancelOk(method)
	case *amqp.BasicPublish:
		return channel.basicPublish(method)
	case *amqp.BasicReturn:
		return channel.basicReturn(method)
	case *amqp.BasicDeliver:
		return channel.basicDeliver(method)
	case *amqp.BasicGet:
		return channel.basicGet(method)
	case *amqp.BasicGetOk:
		return channel.basicGetOk(method)
	case *amqp.BasicGetEmpty:
		return channel.basicGetEmpty(method)
	case *amqp.BasicAck:
		return channel.basicAck(method)
	case *amqp.BasicReject:
		return channel.basicReject(method)
	}
	return errors.New("Unable to route method frame")
}

func (channel *Channel) basicQos(method *amqp.BasicQos) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) basicRecoverAsync(method *amqp.BasicRecoverAsync) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	fmt.Println("Handling BasicRecoverAsync")
	return nil
}

func (channel *Channel) basicQosOk(method *amqp.BasicQosOk) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	fmt.Println("Handling BasicQosOk")
	return nil
}

func (channel *Channel) basicRecover(method *amqp.BasicRecover) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicRecover")
	return nil
}

func (channel *Channel) basicRecoverOk(method *amqp.BasicRecoverOk) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicRecoverOk")
	return nil
}

func (channel *Channel) basicNack(method *amqp.BasicNack) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicNack")
	return nil
}

func (channel *Channel) basicConsume(method *amqp.BasicConsume) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicConsume")
	return nil
}

func (channel *Channel) basicConsumeOk(method *amqp.BasicConsumeOk) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicConsumeOk")
	return nil
}

func (channel *Channel) basicCancel(method *amqp.BasicCancel) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicCancel")
	return nil
}

func (channel *Channel) basicCancelOk(method *amqp.BasicCancelOk) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicCancelOk")
	return nil
}

func (channel *Channel) basicPublish(method *amqp.BasicPublish) error {
	channel.lastMethodFrame = method
	return nil
}

func (channel *Channel) basicReturn(method *amqp.BasicReturn) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicReturn")
	return nil
}

func (channel *Channel) basicDeliver(method *amqp.BasicDeliver) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicDeliver")
	return nil
}

func (channel *Channel) basicGet(method *amqp.BasicGet) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicGet")
	return nil
}

func (channel *Channel) basicGetOk(method *amqp.BasicGetOk) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicGetOk")
	return nil
}

func (channel *Channel) basicGetEmpty(method *amqp.BasicGetEmpty) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicGetEmpty")
	return nil
}

func (channel *Channel) basicAck(method *amqp.BasicAck) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicAck")
	return nil
}

func (channel *Channel) basicReject(method *amqp.BasicReject) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)

	fmt.Println("Handling BasicReject")
	return nil
}
