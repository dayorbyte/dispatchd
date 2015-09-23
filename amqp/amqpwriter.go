package amqp

import (
  "encoding/binary"
  "io"
  "errors"
)

func WriteFrame(buf io.Writer, frame *FrameWrapper) {
  WriteOctet(buf, frame.FrameType)
  WriteShort(buf, frame.Channel)
  WriteLongstr(buf, frame.Payload)
  // buf.Write(frame.Payload.Bytes())
  WriteFrameEnd(buf)
}

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

// func WriteMethodFrameHeader(buf io.Writer, channel uint16, payloadSize uint32) (err error) {
//   if err = binary.Write(buf, binary.BigEndian, FrameTypeMethod); err != nil {
//     return
//   }
//   if err = binary.Write(buf, binary.BigEndian, channel); err != nil {
//     return
//   }
//   if err = binary.Write(buf, binary.BigEndian, payloadSize); err != nil {
//     return
//   }
// }

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
    return binary.Write(buf, binary.BigEndian, 1)
  }
  return binary.Write(buf, binary.BigEndian, 0)

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
  return binary.Write(buf, binary.BigEndian, byte(len(s)))
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

func WriteTable(buf io.Writer, t Table) error {
  // TODO
  return binary.Write(buf, binary.BigEndian, uint32(0))
}
