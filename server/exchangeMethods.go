package main

import (
	_ "fmt"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) exchangeRoute(methodFrame amqp.MethodFrame) *AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.ExchangeDeclare:
		return channel.exchangeDeclare(method)
	case *amqp.ExchangeBind:
		return channel.exchangeBind(method)
	case *amqp.ExchangeUnbind:
		return channel.exchangeUnbind(method)
	case *amqp.ExchangeDelete:
		return channel.exchangeDelete(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) exchangeDeclare(method *amqp.ExchangeDeclare) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	// The client I'm using for testing thought declaring the empty exchange
	// was OK. Check later
	// if len(method.Exchange) > 0 && !method.Passive {
	// 	var msg = "The empty exchange name is reserved"
	// 	channel.channelErrorWithMethod(406, msg, classId, methodId)
	// 	return nil
	// }

	// Check the name format
	var err = amqp.CheckExchangeOrQueueName(method.Exchange)
	if err != nil {
		return NewSoftError(406, err.Error(), classId, methodId)
	}

	// Declare!
	errCode, err := channel.server.declareExchange(method, false, false)
	if err != nil {
		return NewSoftError(errCode, err.Error(), classId, methodId)
	}
	if !method.NoWait {
		channel.SendMethod(&amqp.ExchangeDeclareOk{})
	}
	return nil
}

func (channel *Channel) exchangeDelete(method *amqp.ExchangeDelete) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	var errCode, err = channel.server.deleteExchange(method)
	if err != nil {
		return NewSoftError(errCode, err.Error(), classId, methodId)
	}
	if !method.NoWait {
		channel.SendMethod(&amqp.ExchangeDeleteOk{})
	}
	return nil
}

func (channel *Channel) exchangeBind(method *amqp.ExchangeBind) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	return NewHardError(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.SendMethod(&amqp.ExchangeBindOk{})
	// }
}
func (channel *Channel) exchangeUnbind(method *amqp.ExchangeUnbind) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	return NewHardError(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.SendMethod(&amqp.ExchangeUnbindOk{})
	// }
}
