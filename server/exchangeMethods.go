package main

import (
	"errors"
	"fmt"
	// "bytes"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) exchangeRoute(methodFrame amqp.MethodFrame) error {
	switch method := methodFrame.(type) {
	case *amqp.ExchangeDeclare:
		return channel.exchangeDeclare(method)
	// case *amqp.ExchangeDeclareOk:
	//   return channel.exchangeDeclareOk(method)
	case *amqp.ExchangeDelete:
		return channel.exchangeDelete(method)
		// case *amqp.ExchangeDeleteOk:
		//   return channel.exchangeDeleteOk(method)
	}
	return errors.New("Unable to route method frame")
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
