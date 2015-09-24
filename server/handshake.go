package main

import (
  "fmt"
  "errors"
  "bytes"
  "github.com/jeffjenkins/mq/amqp"
)

func (conn *AMQPConnection) openOk() error {
  fmt.Println("=====> Sending 'openOk'")
  conn.send(&amqp.ConnectionOpenOk{""})
  return conn.openOk()
}

func (conn *AMQPConnection) openStep(startOk *amqp.ConnectionTuneOk) error {
  fmt.Println("=====> Receiving 'open'")
  method, err := conn.receive()
  if err != nil {
    return err
  }
  _, rightType := method.(*amqp.ConnectionOpen)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected open")
  }
  return nil
}

func (conn *AMQPConnection) tuneOk() error {
  fmt.Println("=====> Receiving 'tuneOk'")
  method, err := conn.receive()
  if err != nil {
    return err
  }

  tok, rightType := method.(*amqp.ConnectionTuneOk)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected tunOk")
  }
  return conn.openStep(tok)
}

func (conn *AMQPConnection) tune(startOk *amqp.ConnectionStartOk) error {
  fmt.Println("=====> Sending 'tune'")
  conn.send(&amqp.ConnectionTune{256, 8192, 10})
  return conn.tuneOk()
}

func (conn *AMQPConnection) startOk() error {
  fmt.Println("=====> Receiving 'startOk'")
  method, err := conn.receive()
  if err != nil {
    return err
  }

  sok, rightType := method.(*amqp.ConnectionStartOk)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected startOk")
  }
  return conn.tune(sok)
}

func (conn *AMQPConnection) startConnection() error {
  fmt.Println("=====> Sending 'start'")
  conn.send(&amqp.ConnectionStart{0, 9, amqp.Table{}, []byte("PLAIN"), []byte("en_US")})
  return conn.startOk()
}

func (conn *AMQPConnection) send(method amqp.MethodFrame) error {
  var buf = bytes.NewBuffer([]byte{})
  method.Write(buf)
  var frame = &amqp.FrameWrapper{uint8(amqp.FrameMethod), 0, buf.Bytes()}
  amqp.WriteFrame(conn.network, frame)
  return nil
}

func (conn *AMQPConnection) receive() (amqp.MethodFrame, error) {
  for {
    frame, err := amqp.ReadFrame(conn.network)
    if err != nil {
      return nil, err
    }
    fmt.Println("Frame info", frame.FrameType, frame.Channel)
    var methodReader = bytes.NewReader(frame.Payload)
    return amqp.ReadMethod(methodReader)
  }

}