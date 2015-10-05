package main

import (
	"bytes"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
)

const (
	CH_STATE_INIT    = iota
	CH_STATE_OPEN    = iota
	CH_STATE_CLOSING = iota
	CH_STATE_CLOSED  = iota
)

type Channel struct {
	id              uint16
	server          *Server
	incoming        chan *amqp.WireFrame
	outgoing        chan *amqp.WireFrame
	conn            *AMQPConnection
	state           uint8
	confirmMode     bool
	lastMethodFrame amqp.MethodFrame
	lastHeaderFrame *amqp.ContentHeaderFrame
	bodyFrames      []*amqp.WireFrame
	msgIndex        uint64
	// TODO(MUST): track consumers so they can be cleanly shut down when
	// the channel/connection closes
}

func NewChannel(id uint16, conn *AMQPConnection) *Channel {
	return &Channel{
		id:         id,
		server:     conn.server,
		incoming:   make(chan *amqp.WireFrame),
		outgoing:   conn.outgoing,
		conn:       conn,
		state:      CH_STATE_INIT,
		bodyFrames: make([]*amqp.WireFrame, 0, 1),
	}
}

func (channel *Channel) start() {
	if channel.id == 0 {
		channel.state = CH_STATE_OPEN
		go channel.startConnection()
	} else {
		go channel.startChannel()
	}

	// Receive method frames from the client and route them
	go func() {
		defer channel.destructor()
		for {
			if channel.state == CH_STATE_CLOSED {
				break
			}
			var frame = <-channel.incoming
			switch {
			case frame.FrameType == uint8(amqp.FrameMethod):
				channel.routeMethod(frame)
			case frame.FrameType == uint8(amqp.FrameHeader):
				channel.handleContentHeader(frame)
			case frame.FrameType == uint8(amqp.FrameBody):
				channel.handleContentBody(frame)
			default:
				fmt.Println("Unknown frame type!")
			}
		}
	}()
}

func (channel *Channel) startChannel() {

}

func (channel *Channel) close(code uint16, reason string, classId uint16, methodId uint16) {
	channel.sendMethod(&amqp.ChannelClose{
		ReplyCode: code,
		ReplyText: reason,
		ClassId:   classId,
		MethodId:  methodId,
	})
	channel.state = CH_STATE_CLOSING
}

func (channel *Channel) destructor() {
	channel.conn.deregisterChannel(channel.id)
}

// Send a method frame out to the client
func (channel *Channel) sendMethod(method amqp.MethodFrame) {
	fmt.Printf("Sending method: %s\n", method.MethodName())
	var buf = bytes.NewBuffer([]byte{})
	method.Write(buf)
	channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameMethod), channel.id, buf.Bytes()}
}

// Send a method frame out to the client
func (channel *Channel) sendContent(method *amqp.BasicDeliver, message *Message) {
	fmt.Println("Sending content\n")
	// deliver
	channel.sendMethod(method)
	// header
	channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameHeader), channel.id, message.header.AsBytes}
	// body
	for _, b := range message.payload {
		b.Channel = channel.id
		channel.outgoing <- b
	}
}

func (channel *Channel) handleContentHeader(frame *amqp.WireFrame) {
	if channel.lastMethodFrame == nil {
		fmt.Println("Unexpected content header frame!")
		return
	}
	var headerFrame = &amqp.ContentHeaderFrame{}
	var err = headerFrame.Read(bytes.NewReader(frame.Payload))
	headerFrame.AsBytes = frame.Payload
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
		return
	}
	channel.routeBodyMethod(channel.lastMethodFrame, channel.lastHeaderFrame, channel.bodyFrames)
	channel.lastMethodFrame = nil
	channel.lastHeaderFrame = nil
	channel.bodyFrames = make([]*amqp.WireFrame, 0, 1)
	if channel.confirmMode {
		channel.msgIndex += 1
		channel.sendMethod(&amqp.BasicAck{channel.msgIndex, false})
	}
}

func (channel *Channel) routeMethod(frame *amqp.WireFrame) error {
	var methodReader = bytes.NewReader(frame.Payload)
	var methodFrame, err = amqp.ReadMethod(methodReader)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	var classId, methodId = methodFrame.MethodIdentifier()

	// If the method isn't closing related and we're closing, ignore the frames
	var closeChannel = classId == amqp.ClassIdChannel && (methodId == amqp.MethodIdChannelClose || methodId == amqp.MethodIdChannelCloseOk)
	var closeConnection = classId == amqp.ClassIdConnection && (methodId == amqp.MethodIdConnectionClose || methodId == amqp.MethodIdConnectionCloseOk)
	if channel.state == CH_STATE_CLOSING && !(closeChannel || closeConnection) {
		return nil
	}

	// Route
	fmt.Println("Routing method: " + methodFrame.MethodName())
	switch {
	case classId == 10:
		channel.connectionRoute(methodFrame)
	case classId == 20:
		channel.channelRoute(methodFrame)
	case classId == 40:
		channel.exchangeRoute(methodFrame)
	case classId == 50:
		channel.queueRoute(methodFrame)
	case classId == 60:
		channel.basicRoute(methodFrame)
	case classId == 85:
		channel.confirmRoute(methodFrame)
	default:
		channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	}
	return nil
}

func (channel *Channel) routeBodyMethod(methodFrame amqp.MethodFrame,
	header *amqp.ContentHeaderFrame, bodyFrames []*amqp.WireFrame) {
	switch method := methodFrame.(type) {
	case *amqp.BasicPublish:
		channel.server.exchanges[method.Exchange].publish(method, header, bodyFrames)
	}
}
