package main

import (
	_ "fmt"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) exchangeRoute(methodFrame amqp.MethodFrame) error {
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
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) exchangeDeclare(method *amqp.ExchangeDeclare) error {
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
		channel.channelErrorWithMethod(406, err.Error(), classId, methodId)
		return nil
	}

	if method.Durable {
		panic("Durable not implemented")
	}
	if method.AutoDelete {
		panic("Autodelete not implemented")
	}

	// Declare!
	errCode, err := channel.server.declareExchange(method, false)
	if err != nil {
		channel.channelErrorWithMethod(errCode, err.Error(), classId, methodId)
		return nil
	}
	if !method.NoWait {
		channel.sendMethod(&amqp.ExchangeDeclareOk{})
	}
	return nil
}

func (channel *Channel) exchangeDelete(method *amqp.ExchangeDelete) error {
	var err = channel.server.deleteExchange(method)
	if err != nil {
		return err
	}
	if !method.NoWait {
		channel.sendMethod(&amqp.ExchangeDeleteOk{})
	}
	return nil
}

func (channel *Channel) exchangeBind(method *amqp.ExchangeBind) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.sendMethod(&amqp.ExchangeBindOk{})
	// }
	return nil
}
func (channel *Channel) exchangeUnbind(method *amqp.ExchangeUnbind) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.sendMethod(&amqp.ExchangeUnbindOk{})
	// }
	return nil
}
