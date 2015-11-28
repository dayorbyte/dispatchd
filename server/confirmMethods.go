package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
)

func (channel *Channel) confirmRoute(methodFrame amqp.MethodFrame) *amqp.AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.ConfirmSelect:
		channel.activateConfirmMode()
		return channel.confirmSelect(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return amqp.NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) confirmSelect(method *amqp.ConfirmSelect) *amqp.AMQPError {
	if !method.Nowait {
		channel.SendMethod(&amqp.ConfirmSelectOk{})
	}
	return nil
}
