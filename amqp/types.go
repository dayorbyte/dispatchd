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

type Frame interface {
  FrameType() byte
}

type MethodFrame interface {
  MethodIdentifier() (uint16, uint16)
  Read(reader io.Reader) (err error)
  Write(writer io.Writer) (err error)
  FrameType() byte
}

type WireFrame struct {
  FrameType byte
  Channel uint16
  Payload []byte
}

type ContentFrameHeader interface {
  classId() uint16
  weight() uint16
  bodySize() uint64
}

