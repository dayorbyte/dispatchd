package amqp

import (
  _ "fmt"
  "io"
)

type Decimal struct {
  scale byte
  value int32
}

type Table map[string]interface{}

type MethodFrame interface {
  MethodIdentifier() (uint16, uint16)
  Read(reader io.Reader) (err error)
  Write(writer io.Writer) (err error)
}

type FrameWrapper struct {
  FrameType byte
  Channel uint16
  Payload []byte
}

