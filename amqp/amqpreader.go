package amqp

import (
  "encoding/binary"
  "io"
  "errors"
)

func ReadFrame (reader io.Reader) (*FrameWrapper, error) {
  var f = &FrameWrapper{}
  frameType, err := ReadOctet(reader)
  if err != nil {
    return nil, err
  }
  channel, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  payload, err := ReadLongstr(reader)
  if err != nil {
    return nil, err
  }
  ReadOctet(reader) // Frame end, TODO: assert that this is the right byte.
  f.FrameType = frameType
  f.Channel = channel
  f.Payload = payload
  return f, nil
}

// Constants

func ReadProtocolHeader(buf io.Reader) (err error) {
  var expected = [...]byte {'A', 'M', 'Q', 'P', 0, 0, 9, 1}
  var proto [8]byte
  err = binary.Read(buf, binary.BigEndian, proto)
  if err != nil || proto != expected {
    return errors.New("Bad frame end")
  }
  return nil
}

func ReadVersion(buf io.Reader) (err error) {
  var expected = [2]byte {0, 9}
  var version [2]byte
  err = binary.Read(buf, binary.BigEndian, &version)
  if err != nil || version != expected {
    return errors.New("Bad frame end")
  }
  return nil
}

func ReadFrameEnd(buf io.Reader) (err error) {
  var end byte
  err = binary.Read(buf, binary.BigEndian, &end)
  if err != nil || end != byte(0xCE) {
    return errors.New("Bad frame end")
  }
  return nil
}

// Composite Types

func ReadFrameHeader(buf io.Reader) (frameType byte, channel uint16, payloadSize uint32, err error) {
  binary.Read(buf, binary.BigEndian, &frameType)
  binary.Read(buf, binary.BigEndian, &channel)
  binary.Read(buf, binary.BigEndian, &payloadSize)
  return frameType, channel, payloadSize, nil
}

func ReadMethodPayloadHeader(buf io.Reader) (classId uint16, methodId uint16, err error) {
  binary.Read(buf, binary.BigEndian, &classId)
  binary.Read(buf, binary.BigEndian, &methodId)
  return
}

// Fields

func ReadBit(buf io.Reader) (bool, error) {
  var data byte
  err := binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return false, errors.New("Could not read byte")
  }
  return data != 0, nil
}

func ReadOctet(buf io.Reader) (data byte, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read byte")
  }
  return data, nil
}

func ReadShort(buf io.Reader) (data uint16, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint16")
  }
  return data, nil
}

func ReadLong(buf io.Reader) (data uint32, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint32")
  }
  return data, nil
}

func ReadLonglong(buf io.Reader) (data uint64, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint64")
  }
  return data, nil
}

func ReadStringChar(buf io.Reader) (data byte, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read byte")
  }
  return data, nil
}

func ReadShortstr(buf io.Reader) (string, error) {
  var length uint8
  binary.Read(buf, binary.BigEndian, &length)
  if length > MaxShortStringLength {
    return "", errors.New("String too long for short string")
  }
  var slice = make([]byte, length)
  binary.Read(buf, binary.BigEndian, slice)
  return string(slice), nil
}

func ReadLongstr(buf io.Reader) ([]byte, error) {
  var length uint32
  var err = binary.Read(buf, binary.BigEndian, &length)
  var slice = make([]byte, length)
  binary.Read(buf, binary.BigEndian, slice)
  return slice, err
}

func ReadTimestamp(buf io.Reader) (data uint64, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint64")
  }
  return data, nil
}

