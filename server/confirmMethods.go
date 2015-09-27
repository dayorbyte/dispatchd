
package main

import (
  // "fmt"
  "errors"
  // "bytes"
  "github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) confirmRoute(methodFrame amqp.MethodFrame) error {
  switch method := methodFrame.(type) {
  case *amqp.ConfirmSelect:
    channel.confirmMode = true
    return channel.confirmSelect(method)
    // case *amqp.ConfirmSelectOk:
    //   return channel.confirmSelectOk(method)
  }
  return errors.New("Unable to route method frame")
}


func (channel *Channel) confirmSelect(method *amqp.ConfirmSelect) error {
  channel.sendMethod(&amqp.ConfirmSelectOk{})
  return nil
}
