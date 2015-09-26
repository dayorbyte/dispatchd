package main

import (
  "fmt"
  "bytes"
  "github.com/jeffjenkins/mq/amqp"
  "strconv"
)

type Channel struct {
  id uint16
  incoming chan *amqp.FrameWrapper
  outgoing chan *amqp.FrameWrapper
  conn *AMQPConnection
  open bool
  done bool
}

func NewChannel(id uint16, conn *AMQPConnection) *Channel {
  return &Channel{id, make(chan *amqp.FrameWrapper), conn.outgoing, conn, false, false}
}

func (channel *Channel) start() {
  if channel.id == 0 {
    go channel.startConnection()
  } else {
    channel.open = true
    go channel.startChannel()
  }

  // Receive method frames from the client and route them
  go func() {
    for {
      var frame = <- channel.incoming
      var methodReader = bytes.NewReader(frame.Payload)
      var method, err = amqp.ReadMethod(methodReader)
      if err != nil {
        fmt.Println("ERROR: ", err)
      }
      channel.route(method)
    }
  }()
}

func (channel *Channel) startChannel() {
  fmt.Println("Start channel", channel.id)
}

// Send a method frame out to the client
func (channel *Channel) sendMethod(method amqp.MethodFrame) {
  var buf = bytes.NewBuffer([]byte{})
  method.Write(buf)
  channel.outgoing <- &amqp.FrameWrapper{uint8(amqp.FrameMethod), channel.id, buf.Bytes()}
}

func (channel *Channel) route(methodFrame amqp.MethodFrame) {
  var classId, _ = methodFrame.MethodIdentifier()
  switch {
    case classId == 10:
      channel.connectionRoute(methodFrame)
    case classId == 20:
      channel.channelRoute(methodFrame)
    case classId == 40:
      channel.exchangeRoute(methodFrame)
    default:
      panic("Not implemented! " + strconv.FormatUint(uint64(classId), 10))
  }
}
