package amqp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

func WriteFrame(buf io.Writer, frame *WireFrame) {
	bb := make([]byte, 0, 1+2+4+len(frame.Payload)+2)
	buf2 := bytes.NewBuffer(bb)
	WriteOctet(buf2, frame.FrameType)
	WriteShort(buf2, frame.Channel)
	WriteLongstr(buf2, frame.Payload)
	// buf.Write(frame.Payload.Bytes())
	WriteFrameEnd(buf2)
	// binary.LittleEndian since we want to stick to the system
	// byte order and the other write functions are writing BigEndian
	binary.Write(buf, binary.LittleEndian, buf2.Bytes())
}

// Constants

func WriteProtocolHeader(buf io.Writer) error {
	return binary.Write(buf, binary.BigEndian, []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1})
}

func WriteVersion(buf io.Writer) error {
	return binary.Write(buf, binary.BigEndian, []byte{0, 9})
}

func WriteFrameEnd(buf io.Writer) error {
	return binary.Write(buf, binary.BigEndian, byte(0xCE))
}

// Fields

func WriteOctet(buf io.Writer, b byte) error {
	return binary.Write(buf, binary.BigEndian, b)
}

func WriteShort(buf io.Writer, i uint16) error {
	return binary.Write(buf, binary.BigEndian, i)
}

func WriteLong(buf io.Writer, i uint32) error {
	return binary.Write(buf, binary.BigEndian, i)
}

func WriteLonglong(buf io.Writer, i uint64) error {
	return binary.Write(buf, binary.BigEndian, i)
}

func WriteStringChar(buf io.Writer, b byte) error {
	return binary.Write(buf, binary.BigEndian, b)
}

func WriteShortstr(buf io.Writer, s string) error {
	if len(s) > int(MaxShortStringLength) {
		return errors.New("String too long for short string")
	}
	err := binary.Write(buf, binary.BigEndian, byte(len(s)))
	if err != nil {
		return errors.New("Could not write bytes: " + err.Error())
	}
	return binary.Write(buf, binary.BigEndian, []byte(s))
}

func WriteLongstr(buf io.Writer, bytes []byte) (err error) {
	if err = binary.Write(buf, binary.BigEndian, uint32(len(bytes))); err != nil {
		return
	}
	if err = binary.Write(buf, binary.BigEndian, bytes); err != nil {
		return
	}
	return nil
}

func WriteTimestamp(buf io.Writer, timestamp uint64) error {
	return binary.Write(buf, binary.BigEndian, timestamp)
}

func WriteTable(writer io.Writer, table *Table) error {
	var buf = bytes.NewBuffer(make([]byte, 0))
	for _, kv := range table.Table {
		if err := WriteShortstr(buf, *kv.Key); err != nil {
			return err
		}
		if err := writeValue(buf, kv.Value); err != nil {
			return err
		}
	}
	return WriteLongstr(writer, buf.Bytes())
}

func writeArray(writer io.Writer, array []*FieldValue) error {
	var buf = bytes.NewBuffer([]byte{})
	for _, v := range array {
		if err := writeValue(buf, v); err != nil {
			return err
		}
	}
	return WriteLongstr(writer, buf.Bytes())
}

func writeValue(writer io.Writer, value *FieldValue) (err error) {
	switch v := value.Value.(type) {
	case *FieldValue_VBoolean:
		if err = binary.Write(writer, binary.BigEndian, byte('t')); err == nil {
			if v.VBoolean {
				err = WriteOctet(writer, uint8(1))
			} else {
				err = WriteOctet(writer, uint8(0))
			}
		}
	case *FieldValue_VInt8:
		if err = binary.Write(writer, binary.BigEndian, byte('b')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int8(v.VInt8))
		}
	case *FieldValue_VUint8:
		if err = binary.Write(writer, binary.BigEndian, byte('B')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint8(v.VUint8))
		}
	case *FieldValue_VInt16:
		if err = binary.Write(writer, binary.BigEndian, byte('U')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int16(v.VInt16))
		}
	case *FieldValue_VUint16:
		if err = binary.Write(writer, binary.BigEndian, byte('u')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint16(v.VUint16))
		}
	case *FieldValue_VInt32:
		if err = binary.Write(writer, binary.BigEndian, byte('I')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int32(v.VInt32))
		}
	case *FieldValue_VUint32:
		if err = binary.Write(writer, binary.BigEndian, byte('i')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint32(v.VUint32))
		}
	case *FieldValue_VInt64:
		if err = binary.Write(writer, binary.BigEndian, byte('L')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int64(v.VInt64))
		}
	case *FieldValue_VUint64:
		if err = binary.Write(writer, binary.BigEndian, byte('l')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint64(v.VUint64))
		}
	case *FieldValue_VFloat:
		if err = binary.Write(writer, binary.BigEndian, byte('f')); err == nil {
			err = binary.Write(writer, binary.BigEndian, float32(v.VFloat))
		}
	case *FieldValue_VDouble:
		if err = binary.Write(writer, binary.BigEndian, byte('d')); err == nil {
			err = binary.Write(writer, binary.BigEndian, float64(v.VDouble))
		}
	case *FieldValue_VDecimal:
		if err = binary.Write(writer, binary.BigEndian, byte('D')); err == nil {
			if err = binary.Write(writer, binary.BigEndian, byte(*v.VDecimal.Scale)); err == nil {
				err = binary.Write(writer, binary.BigEndian, uint32(*v.VDecimal.Value))
			}
		}
	case *FieldValue_VShortstr:
		if err = WriteOctet(writer, byte('s')); err == nil {
			err = WriteShortstr(writer, v.VShortstr)
		}
	case *FieldValue_VLongstr:
		if err = WriteOctet(writer, byte('S')); err == nil {
			err = WriteLongstr(writer, v.VLongstr)
		}
	case *FieldValue_VArray:
		if err = WriteOctet(writer, byte('A')); err == nil {
			err = writeArray(writer, v.VArray.Value)
		}
	case *FieldValue_VTimestamp:
		if err = WriteOctet(writer, byte('T')); err == nil {
			err = WriteTimestamp(writer, v.VTimestamp)
		}
	case *FieldValue_VTable:
		if err = WriteOctet(writer, byte('F')); err == nil {
			err = WriteTable(writer, v.VTable)
		}
	case nil:
		err = binary.Write(writer, binary.BigEndian, byte('V'))
	default:
		panic("unsupported type!")
	}
	return
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
