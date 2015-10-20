package amqp

import (
	"errors"
	"fmt"
	"io"
	"regexp"
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
	MethodName() string
	MethodIdentifier() (uint16, uint16)
	Read(reader io.Reader) (err error)
	Write(writer io.Writer) (err error)
	FrameType() byte
}

type WireFrame struct {
	FrameType byte
	Channel   uint16
	Payload   []byte
}

type ContentHeaderFrame struct {
	ContentClass    uint16
	ContentWeight   uint16
	ContentBodySize uint64
	PropertyFlags   uint16
	Properties      *BasicContentHeaderProperties
	AsBytes         []byte
}

func NewTruncatedBodyFrame(channel uint16) WireFrame {
	return WireFrame{
		FrameType: byte(FrameBody),
		Channel:   channel,
		Payload:   make([]byte, 0, 0),
	}
}

func (frame *ContentHeaderFrame) FrameType() byte {
	return 2
}

var exchangeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.:]*$`)

func CheckExchangeName(s string) error {
	if len(s) > 127 {
		return fmt.Errorf("Exchange name too long: %d", len(s))
	}
	if !exchangeNameRegex.MatchString(s) {
		return fmt.Errorf("Exchange name invalid: %s", s)
	}
	return nil
}

func (frame *ContentHeaderFrame) Read(reader io.Reader) (err error) {
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
	err = frame.Properties.readProps(frame.PropertyFlags, reader)
	if err != nil {
		return err
	}
	return nil
}

func (props *BasicContentHeaderProperties) readProps(flags uint16, reader io.Reader) (err error) {
	if MaskContentType&flags != 0 {
		props.ContentType, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskContentEncoding&flags != 0 {
		props.ContentEncoding, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskHeaders&flags != 0 {
		props.Headers, err = ReadTable(reader)
		if err != nil {
			return
		}
	}
	if MaskDeliveryMode&flags != 0 {
		props.DeliveryMode, err = ReadOctet(reader)
		if err != nil {
			return
		}
	}
	if MaskPriority&flags != 0 {
		props.Priority, err = ReadOctet(reader)
		if err != nil {
			return
		}
	}
	if MaskCorrelationId&flags != 0 {
		props.CorrelationId, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskReplyTo&flags != 0 {
		props.ReplyTo, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskExpiration&flags != 0 {
		props.Expiration, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskMessageId&flags != 0 {
		props.MessageId, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskTimestamp&flags != 0 {
		props.Timestamp, err = ReadLonglong(reader)
		if err != nil {
			return
		}
	}
	if MaskType&flags != 0 {
		props.Type, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskUserId&flags != 0 {
		props.UserId, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskAppId&flags != 0 {
		props.AppId, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	if MaskReserved&flags != 0 {
		props.Reserved, err = ReadShortstr(reader)
		if err != nil {
			return
		}
	}
	return nil
}
