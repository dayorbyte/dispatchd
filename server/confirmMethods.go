package main

import (
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) confirmRoute(methodFrame amqp.MethodFrame) *AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.ConfirmSelect:
		channel.activateConfirmMode()
		return channel.confirmSelect(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) confirmSelect(method *amqp.ConfirmSelect) *AMQPError {
	if !method.Nowait {
		channel.SendMethod(&amqp.ConfirmSelectOk{})
	}
	return nil
}
