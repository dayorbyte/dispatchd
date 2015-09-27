package main

import (
  "fmt"
  "bytes"
  "github.com/jeffjenkins/mq/amqp"
  "strconv"
)

type Channel struct {
  id uint16
  server *Server
  incoming chan *amqp.WireFrame
  outgoing chan *amqp.WireFrame
  conn *AMQPConnection
  open bool
  done bool
  lastMethodFrame amqp.MethodFrame
  lastHeaderFrame *amqp.ContentHeaderFrame
  bodyFrames []*amqp.WireFrame

}

func NewChannel(id uint16, conn *AMQPConnection) *Channel {
  return &Channel{
    id,
    conn.server,
    make(chan *amqp.WireFrame),
    conn.outgoing,
    conn,
    false,
    false,
    nil,
    nil,
    make([]*amqp.WireFrame, 0, 1),
  }
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
      switch {
      case frame.FrameType == 1:
        channel.routeMethod(frame)
      case frame.FrameType == 2:
        channel.handleContentHeader(frame)
      case frame.FrameType == 3:
        channel.handleContentBody(frame)
      default:
        fmt.Println("Unknown frame type!")
      }
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
  channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameMethod), channel.id, buf.Bytes()}
}

func (channel *Channel) handleContentHeader(frame *amqp.WireFrame) {
  if channel.lastMethodFrame == nil {
    fmt.Println("Unexpected content header frame!")
    return
  }
  var headerFrame = &amqp.ContentHeaderFrame{}
  var err = headerFrame.Read(bytes.NewReader(frame.Payload))
  if err != nil {
    fmt.Println("Error parsing header frame: " + err.Error())
  }
  channel.lastHeaderFrame = headerFrame
}

func (channel *Channel) handleContentBody(frame *amqp.WireFrame) {
  if channel.lastMethodFrame == nil {
    fmt.Println("Unexpected content body frame. No method content-having method called yet!")
  }
  if channel.lastHeaderFrame == nil {
    fmt.Println("Unexpected content body frame! No header yet")
  }
  channel.bodyFrames = append(channel.bodyFrames, frame)
  var size = uint64(0)
  for _, body := range channel.bodyFrames {
    size += uint64(len(body.Payload))
  }
  if size < channel.lastHeaderFrame.ContentBodySize {
    fmt.Println("More body to come!")
    return
  }
  fmt.Println("Dispatching")
  channel.routeBodyMethod(channel.lastMethodFrame, channel.lastHeaderFrame, channel.bodyFrames)

}

func (channel *Channel) routeMethod(frame *amqp.WireFrame) error {
  var methodReader = bytes.NewReader(frame.Payload)
  var methodFrame, err = amqp.ReadMethod(methodReader)
  if err != nil {
    fmt.Println("ERROR: ", err)
  }
  var classId, _ = methodFrame.MethodIdentifier()
  switch {
    case classId == 10:
      channel.connectionRoute(methodFrame)
    case classId == 20:
      channel.channelRoute(methodFrame)
    case classId == 40:
      channel.exchangeRoute(methodFrame)
    case classId == 60:
      channel.basicRoute(methodFrame)
    case classId == 85:
      channel.confirmRoute(methodFrame)
    default:
      panic("Not implemented! " + strconv.FormatUint(uint64(classId), 10))
  }
  return nil
}

func (channel *Channel) routeBodyMethod(methodFrame amqp.MethodFrame, header *amqp.ContentHeaderFrame, bodyFrames []*amqp.WireFrame) {
  switch method := methodFrame.(type) {
  case *amqp.BasicPublish:
    channel.server.exchanges[method.Exchange].publish(method, header, bodyFrames)
  }
}