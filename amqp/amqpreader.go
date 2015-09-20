package amqp

import (
  "encoding/binary"
  "io"
  "errors"
  "bytes"
)

// Constants

func ReadProtocolHeader(buf io.Reader) error {
  var expected = [...]byte {'A', 'M', 'Q', 'P', 0, 0, 9, 1}
  var proto [8]byte
  err = binary.Read(buf, binary.BigEndian, proto)
  if err != nil || proto != expected {
    return errors.New("Bad frame end")
  }
  return nil
}

func ReadVersion(buf io.Reader) error {
  expected = [2]byte {0, 9}
  var version [2]byte
  err = binary.Read(buf, binary.BigEndian, &version)
  if err != nil || version != expected {
    return errors.New("Bad frame end")
  }
  return nil
}

func ReadFrameEnd(buf io.Reader) error {
  var end byte
  err = binary.Read(buf, binary.BigEndian, &end)
  if err != nil || end != byte(0xCE) {
    return errors.New("Bad frame end")
  }
  return nil
}

// Composite Types

func ReadFrameHeader(buf io.Reader) byte, uint16, uint32, error {
  var channel uint16
  var payloadSize uint32
  var frameType byte
  binary.Read(buf, binary.BigEndian, &frameType)
  binary.Read(buf, binary.BigEndian, &channel)
  binary.Read(buf, binary.BigEndian, &payloadSize)
  return frameType, channel, payloadSize, nil
}

func ReadMethodPayloadHeader(buf io.Reader) uint16, uint16, error {
  var classId uint16
  var methodId uint16
  binary.Read(buf, binary.BigEndian, &classId)
  binary.Read(buf, binary.BigEndian, &methodId)
}

// Fields

func ReadBits(buf io.Reader) []bool, error{
  panic("Not implemented!")
}

func ReadOctet(buf io.Reader) byte, error {
  var in byte
  err = binary.Read(buf, binary.BigEndian, &in)
  if err != nil {
    return nil, errors.New("Could not read byte")
  }
  return in, nil
}

func ReadShort(buf io.Reader) uint16, error {
  var in uint16
  err = binary.Read(buf, binary.BigEndian, &in)
  if err != nil {
    return nil, errors.New("Could not read uint16")
  }
  return in, nil
}

func ReadLong(buf io.Reader) uint32, error {
  var in uint32
  err = binary.Read(buf, binary.BigEndian, &in)
  if err != nil {
    return nil, errors.New("Could not read uint32")
  }
  return in, nil
}

func ReadLongLong(buf io.Reader) uint64, error {
  var in uint64
  err = binary.Read(buf, binary.BigEndian, &in)
  if err != nil {
    return nil, errors.New("Could not read uint64")
  }
  return in, nil
}

func ReadStringChar(buf io.Reader) byte, error {
  var in byte
  err = binary.Read(buf, binary.BigEndian, &in)
  if err != nil {
    return nil, errors.New("Could not read byte")
  }
  return in, nil
}

func ReadShortString(buf io.Reader) string, error {
  var length uint8
  err = binary.Read(buf, binary.BigEndian, &length)
  if length > MaxShortStringLength {
    return nil, errors.New("String too long for short string")
  }
  var slice = new([]byte, length)
  binary.Read(buf, binary.BigEndian, slice)
  return string(slice), nil
}

func ReadLongString(buf io.Reader) []byte, error {
  var length uint32
  err = binary.Read(buf, binary.BigEndian, &length)
  if length > MaxShortStringLength {
    return nil, errors.New("String too long for short string")
  }
  var slice = new([]byte, length)
  binary.Read(buf, binary.BigEndian, slice)
  return slice
}

func ReadTimestamp(buf io.Reader) uint64, error {
  var in uint64
  err = binary.Read(buf, binary.BigEndian, &in)
  if err != nil {
    return nil, errors.New("Could not read uint64")
  }
  return in, nil
}

func ReadTable(buf io.Reader) {
  panic("Not implemented")
  // binary.Write(buf, binary.BigEndian, 0)
}

func ReadPeerProperties(buf io.Reader) {
  ReadTable(buf)
}