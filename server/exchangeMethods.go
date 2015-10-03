package main

import (
	"fmt"
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
	var err = channel.server.declareExchange(method)
	if err != nil {
		fmt.Println("Declaring exchange: Error")
		// TODO(send error to client)

		return err
	}
	fmt.Println("Declaring exchange: Send Response")
	channel.sendMethod(&amqp.ExchangeDeclareOk{})
	return nil
}

func (channel *Channel) exchangeDelete(method *amqp.ExchangeDelete) error {
	var err = channel.server.deleteExchange(method)
	if err != nil {
		return err
	}
	channel.sendMethod(&amqp.ExchangeDeleteOk{})
	return nil
}

func (channel *Channel) exchangeBind(method *amqp.ExchangeBind) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}
func (channel *Channel) exchangeUnbind(method *amqp.ExchangeUnbind) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}
