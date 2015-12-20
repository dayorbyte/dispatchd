package amqp

import (
	"errors"
	"fmt"
	"github.com/jeffjenkins/dispatchd/util"
	"io"
	"regexp"
)

type Frame interface {
	FrameType() byte
}

type MethodFrame interface {
	MethodName() string
	MethodIdentifier() (uint16, uint16)
	Read(reader io.Reader, strictMode bool) (err error)
	Write(writer io.Writer) (err error)
	FrameType() byte
}

// A message resource is something which has limits on the count of
// messages it can handle as well as the cumulative size of the messages
// it can handle.
type MessageResourceHolder interface {
	AcquireResources(qm *QueueMessage) bool
	ReleaseResources(qm *QueueMessage)
}

func NewMessage(method *BasicPublish, localId int64) *Message {
	return &Message{
		Id:       util.NextId(),
		Method:   method,
		Exchange: method.Exchange,
		Key:      method.RoutingKey,
		Payload:  make([]*WireFrame, 0, 1),
		LocalId:  localId,
	}
}

func NewTruncatedBodyFrame(channel uint16) WireFrame {
	return WireFrame{
		FrameType: byte(FrameBody),
		Channel:   channel,
		Payload:   make([]byte, 0, 0),
	}
}

func NewUnackedMessage(tag string, qm *QueueMessage, queueName string) *UnackedMessage {
	return &UnackedMessage{
		ConsumerTag: tag,
		Msg:         qm,
		QueueName:   queueName,
	}
}

func NewIndexMessage(id int64, refCount int32, durable bool, deliveryCount int32) *IndexMessage {
	return &IndexMessage{
		Id:            id,
		Refs:          refCount,
		Durable:       durable,
		DeliveryCount: deliveryCount,
	}
}

func NewQueueMessage(id int64, deliveryCount int32, durable bool, msgSize uint32, localId int64) *QueueMessage {
	return &QueueMessage{
		Id:            id,
		DeliveryCount: deliveryCount,
		Durable:       durable,
		MsgSize:       msgSize,
		LocalId:       localId,
		// NOTE: When loading this from disk later we should zero localId out. This is the ID
		//       of the publishing channel, which isn't relevant on server boot.
	}
}

func (frame *ContentHeaderFrame) FrameType() byte {
	return 2
}

var exchangeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.:]*$`)

func CheckExchangeOrQueueName(s string) error {
	// Is it possible this length check is generally ignored since a short
	// string is only twice as long?
	if len(s) > 127 {
		return fmt.Errorf("Exchange name too long: %d", len(s))
	}
	if !exchangeNameRegex.MatchString(s) {
		return fmt.Errorf("Exchange name invalid: %s", s)
	}
	return nil
}

func (frame *ContentHeaderFrame) Read(reader io.Reader, strictMode bool) (err error) {
	frame.ContentClass, err = ReadShort(reader)
	if err != nil {
		return err
	}

	frame.ContentWeight, err = ReadShort(reader)
	if err != nil {
		return err
	}
	if frame.ContentWeight != 0 {
		return errors.New("Bad content weight in header frame. Should be 0")
	}

	frame.ContentBodySize, err = ReadLonglong(reader)
	if err != nil {
		return err
	}

	frame.PropertyFlags, err = ReadShort(reader)
	if err != nil {
		return err
	}

	frame.Properties = &BasicContentHeaderProperties{}
	err = frame.Properties.ReadProps(frame.PropertyFlags, reader, strictMode)
	if err != nil {
		return err
	}
	return nil
}
