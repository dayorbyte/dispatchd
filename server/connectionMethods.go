package main

import (
  "fmt"
  "errors"
  // "bytes"
  "github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) connectionRoute(methodFrame amqp.MethodFrame) error {
  switch method := methodFrame.(type) {
  case *amqp.ConnectionStartOk:
    return channel.connectionStartOk(method)
  case *amqp.ConnectionTuneOk:
    return channel.connectionTuneOk(method)
  case *amqp.ConnectionOpen:
    return channel.connectionOpen(method)
  }
  return errors.New("Unable to route method frame")
}

func (channel *Channel) connectionOpen(method *amqp.ConnectionOpen) error {
  channel.conn.connectStatus.open = true
  fmt.Println("=====> Sending 'openOk'")
  channel.sendMethod(&amqp.ConnectionOpenOk{""})
  channel.conn.connectStatus.openOk = true
  channel.conn.open = true
  return nil
}

func (channel *Channel) connectionTuneOk(method *amqp.ConnectionTuneOk) error {
  // TODO
  channel.conn.connectStatus.tuneOk = true
  return nil
}

func (channel *Channel) connectionStartOk(method *amqp.ConnectionStartOk) error {
  channel.conn.connectStatus.startOk = true
  fmt.Println("=====> Sending 'tune'")
  channel.sendMethod(&amqp.ConnectionTune{256, 8192, 10})
  channel.conn.connectStatus.tune = true
  return nil
}

func (channel *Channel) startConnection() error {
  fmt.Println("=====> Sending 'start'")
  channel.sendMethod(&amqp.ConnectionStart{0, 9, amqp.Table{}, []byte("PLAIN"), []byte("en_US")})
  return nil
}
