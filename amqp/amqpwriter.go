package amqp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"
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

// Composite Types

func WriteMethodPayloadHeader(buf io.Writer, classId uint16, methodId uint16) (err error) {
	if err = binary.Write(buf, binary.BigEndian, classId); err != nil {
		return
	}
	if err = binary.Write(buf, binary.BigEndian, methodId); err != nil {
		return
	}
	return nil
}

// Fields

func WriteBit(buf io.Writer, b bool) error {
	// TODO: pack these properly
	if b {
		return binary.Write(buf, binary.BigEndian, byte(1))
	}
	return binary.Write(buf, binary.BigEndian, byte(0))

}

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

func WriteTimestamp(buf io.Writer, timestamp time.Time) error {
	return binary.Write(buf, binary.BigEndian, uint64(timestamp.Unix()))
}

func WriteTable(writer io.Writer, table *Table) error {
	var buf = bytes.NewBuffer([]byte{})
	for _, kv := range table.Table {
		WriteShortstr(buf, *kv.Key)
		writeValue(buf, kv.Value.Value)
	}
	return WriteLongstr(writer, buf.Bytes())
}

func writeArray(writer io.Writer, array []interface{}) error {
	var buf = bytes.NewBuffer([]byte{})
	for _, v := range array {
		if err := writeValue(buf, v); err != nil {
			return err
		}
	}
	return WriteLongstr(writer, buf.Bytes())
}

func writeValue(writer io.Writer, value interface{}) (err error) {
	switch v := value.(type) {
	case bool:
		if err = binary.Write(writer, binary.BigEndian, byte('t')); err == nil {
			err = WriteBit(writer, v)
		}
	case int8:
		if err = binary.Write(writer, binary.BigEndian, byte('b')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int8(v))
		}
	case uint8:
		if err = binary.Write(writer, binary.BigEndian, byte('B')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint8(v))
		}
	case int16:
		if err = binary.Write(writer, binary.BigEndian, byte('U')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int16(v))
		}
	case uint16:
		if err = binary.Write(writer, binary.BigEndian, byte('u')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint16(v))
		}
	case int32:
		if err = binary.Write(writer, binary.BigEndian, byte('I')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int32(v))
		}
	case uint32:
		if err = binary.Write(writer, binary.BigEndian, byte('i')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint32(v))
		}
	case int64:
		if err = binary.Write(writer, binary.BigEndian, byte('L')); err == nil {
			err = binary.Write(writer, binary.BigEndian, int64(v))
		}
	case uint64:
		if err = binary.Write(writer, binary.BigEndian, byte('l')); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint64(v))
		}
	case float32:
		if err = binary.Write(writer, binary.BigEndian, byte('f')); err == nil {
			err = binary.Write(writer, binary.BigEndian, float32(v))
		}
	case float64:
		if err = binary.Write(writer, binary.BigEndian, byte('d')); err == nil {
			err = binary.Write(writer, binary.BigEndian, float64(v))
		}
	case Decimal:
		if err = binary.Write(writer, binary.BigEndian, byte(*v.Scale)); err == nil {
			err = binary.Write(writer, binary.BigEndian, uint32(*v.Value))
		}
	case string:
		if err = WriteOctet(writer, byte('s')); err == nil {
			err = WriteShortstr(writer, v)
		}
	case []byte:
		if err = WriteOctet(writer, byte('S')); err == nil {
			err = WriteLongstr(writer, v)
		}
	case []interface{}:
		if err = WriteOctet(writer, byte('A')); err == nil {
			err = writeArray(writer, v)
		}
	case time.Time:
		if err = WriteOctet(writer, byte('T')); err == nil {
			err = WriteTimestamp(writer, v)
		}
	case Table:
		if err = WriteOctet(writer, byte('F')); err == nil {
			err = WriteTable(writer, &v)
		}
	case nil:
		err = binary.Write(writer, binary.BigEndian, byte('V'))
	default:
		panic("unsupported type!")
	}
	return
}
