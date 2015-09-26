
package main

import (
  // "fmt"
  "errors"
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
  channel.sendMethod(&amqp.ExchangeDeclareOk{})
  return nil
}

func (channel *Channel) exchangeDelete(method *amqp.ExchangeDelete) error {
  channel.sendMethod(&amqp.ExchangeDeleteOk{})
  return nil
}
