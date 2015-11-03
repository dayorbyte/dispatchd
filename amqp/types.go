package amqp

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
)

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

// type ContentHeaderFrame struct {
// 	ContentClass    uint16
// 	ContentWeight   uint16
// 	ContentBodySize uint64
// 	PropertyFlags   uint16
// 	Properties      *BasicContentHeaderProperties
// 	AsBytes         []byte
// }

func NewTruncatedBodyFrame(channel uint16) WireFrame {
	return WireFrame{
		FrameType: byte(FrameBody),
		Channel:   channel,
		Payload:   make([]byte, 0, 0),
	}
}

func NewTable() *Table {
	return &Table{Table: make([]*FieldValuePair, 0)}
}

func (frame *ContentHeaderFrame) FrameType() byte {
	return 2
}

func EquivalentTables(t1 *Table, t2 *Table) bool {
	return reflect.DeepEqual(t1, t2)
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
	err = frame.Properties.ReadProps(frame.PropertyFlags, reader)
	if err != nil {
		return err
	}
	return nil
}

func (props *BasicContentHeaderProperties) ReadProps(flags uint16, reader io.Reader) (err error) {
	if MaskContentType&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.ContentType = v
		if err != nil {
			return err
		}
	}
	if MaskContentEncoding&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.ContentEncoding = v
		if err != nil {
			return err
		}
	}
	if MaskHeaders&flags != 0 {
		v, err := ReadTable(reader)
		props.Headers = v
		if err != nil {
			return err
		}
	}
	if MaskDeliveryMode&flags != 0 {
		v, err := ReadOctet(reader)
		*props.DeliveryMode = v
		if err != nil {
			return err
		}
	}
	if MaskPriority&flags != 0 {
		v, err := ReadOctet(reader)
		*props.Priority = v
		if err != nil {
			return err
		}
	}
	if MaskCorrelationId&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.CorrelationId = v
		if err != nil {
			return err
		}
	}
	if MaskReplyTo&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.ReplyTo = v
		if err != nil {
			return err
		}
	}
	if MaskExpiration&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.Expiration = v
		if err != nil {
			return err
		}
	}
	if MaskMessageId&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.MessageId = v
		if err != nil {
			return err
		}
	}
	if MaskTimestamp&flags != 0 {
		v, err := ReadLonglong(reader)
		*props.Timestamp = v
		if err != nil {
			return err
		}
	}
	if MaskType&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.Type = v
		if err != nil {
			return err
		}
	}
	if MaskUserId&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.UserId = v
		if err != nil {
			return err
		}
	}
	if MaskAppId&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.AppId = v
		if err != nil {
			return err
		}
	}
	if MaskReserved&flags != 0 {
		v, err := ReadShortstr(reader)
		*props.Reserved = v
		if err != nil {
			return err
		}
	}
	return nil
}

func (props *BasicContentHeaderProperties) WriteProps(writer io.Writer) (flags uint16, err error) {
	if props.ContentType != nil {
		flags = flags | MaskContentType
		err = WriteShortstr(writer, *props.ContentType)
		if err != nil {
			return
		}
	}
	if props.ContentEncoding != nil {
		flags = flags | MaskContentEncoding
		err = WriteShortstr(writer, *props.ContentEncoding)
		if err != nil {
			return
		}
	}
	if props.Headers != nil {
		flags = flags | MaskHeaders
		err = WriteTable(writer, props.Headers)
		if err != nil {
			return
		}
	}
	if props.DeliveryMode != nil {
		flags = flags | MaskDeliveryMode
		err = WriteOctet(writer, *props.DeliveryMode)
		if err != nil {
			return
		}
	}
	if props.Priority != nil {
		flags = flags | MaskPriority
		err = WriteOctet(writer, *props.Priority)
		if err != nil {
			return
		}
	}
	if props.CorrelationId != nil {
		flags = flags | MaskCorrelationId
		err = WriteShortstr(writer, *props.CorrelationId)
		if err != nil {
			return
		}
	}
	if props.ReplyTo != nil {
		flags = flags | MaskReplyTo
		err = WriteShortstr(writer, *props.ReplyTo)
		if err != nil {
			return
		}
	}
	if props.Expiration != nil {
		flags = flags | MaskExpiration
		err = WriteShortstr(writer, *props.Expiration)
		if err != nil {
			return
		}
	}
	if props.MessageId != nil {
		flags = flags | MaskMessageId
		err = WriteShortstr(writer, *props.MessageId)
		if err != nil {
			return
		}
	}
	if props.Timestamp != nil {
		flags = flags | MaskTimestamp
		err = WriteLonglong(writer, *props.Timestamp)
		if err != nil {
			return
		}
	}
	if props.Type != nil {
		flags = flags | MaskType
		err = WriteShortstr(writer, *props.Type)
		if err != nil {
			return
		}
	}
	if props.UserId != nil {
		flags = flags | MaskUserId
		err = WriteShortstr(writer, *props.UserId)
		if err != nil {
			return
		}
	}
	if props.AppId != nil {
		flags = flags | MaskAppId
		err = WriteShortstr(writer, *props.AppId)
		if err != nil {
			return
		}
	}
	if props.Reserved != nil {
		flags = flags | MaskReserved
		err = WriteShortstr(writer, *props.Reserved)
		if err != nil {
			return
		}
	}
	return
}
