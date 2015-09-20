package amqp

import (
  "encoding/binary"
  "io"
  "errors"
)

// Constants

func WriteProtocolHeader(buf io.Writer) error {
  return binary.Write(buf, binary.BigEndian, []byte {'A', 'M', 'Q', 'P', 0, 0, 9, 1})
}

func WriteVersion(buf io.Writer) error {
  return binary.Write(buf, binary.BigEndian, []byte {0, 9})
}

func WriteFrameEnd(buf io.Writer) error {
  return binary.Write(buf, binary.BigEndian, byte(0xCE))
}

// Composite Types

func WriteMethodFrameHeader(buf io.Writer, channel uint16, payloadSize uint32) (err error) {
  if err := binary.Write(buf, binary.BigEndian, FrameTypeMethod); err == nil {
    return
  }
  if err := binary.Write(buf, binary.BigEndian, channel); err == nil {
    return
  }
  if err := binary.Write(buf, binary.BigEndian, payloadSize); err == nil {
    return
  }
  return nil
}

func WriteMethodPayloadHeader(buf io.Writer, classId uint16, methodId uint16) (err error) {
  if err := binary.Write(buf, binary.BigEndian, classId); err == nil {
    return
  }
  if err := binary.Write(buf, binary.BigEndian, methodId); err == nil {
    return
  }
  return nil
}

// Fields

func WriteBits(buf io.Writer, bits []bool) err {
  // TODO
  panic("Not implemented!")
  return nil
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

func WriteLongLong(buf io.Writer, i uint64) error {
  return binary.Write(buf, binary.BigEndian, )
}

func WriteStringChar(buf io.Writer, b byte) error {
  return binary.Write(buf, binary.BigEndian, b)
}

func WriteShortString(buf io.Writer, s string) error {
  if len(s) > MaxShortStringLength {
    return errors.New("String too long for short string")
  }
  binary.Write(buf, binary.BigEndian, byte(len(s)))
}

func WriteLongString(buf io.Writer, bytes []byte) (err error) {
  if err := binary.Write(buf, binary.BigEndian, uint32(len(bytes))); err == nil {
    return
  }
  if err := binary.Write(buf, binary.BigEndian, bytes); err == nil {
    return
  }
  return nil
}

func WriteTimestamp(buf io.Writer, timestamp uint64) error {
  return binary.Write(buf, binary.BigEndian, timestamp)
}

func WriteTable(buf io.Writer) error {
  // TODO
  return binary.Write(buf, binary.BigEndian, 0)
}