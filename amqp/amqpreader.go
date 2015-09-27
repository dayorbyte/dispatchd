package amqp

import (
  "encoding/binary"
  "io"
  "errors"
  "bytes"
  "time"
  // "strconv"
)

func ReadFrame (reader io.Reader) (*WireFrame, error) {
  var f = &WireFrame{}
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
    return 0, errors.New("Could not read byte: " + err.Error())
  }
  return data, nil
}

func ReadShort(buf io.Reader) (data uint16, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint16: " + err.Error())
  }
  return data, nil
}

func ReadLong(buf io.Reader) (data uint32, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint32: " + err.Error())
  }
  return data, nil
}

func ReadLonglong(buf io.Reader) (data uint64, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read uint64: " + err.Error())
  }
  return data, nil
}

func ReadStringChar(buf io.Reader) (data byte, err error) {
  err = binary.Read(buf, binary.BigEndian, &data)
  if err != nil {
    return 0, errors.New("Could not read byte: " + err.Error())
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

func ReadTimestamp(buf io.Reader) (time.Time, error) {
  var t uint64
  var err = binary.Read(buf, binary.BigEndian, &t)
  if err != nil {
    return time.Unix(0, 0), errors.New("Could not read uint64")
  }
  return time.Unix(int64(t), 0), nil
}

func ReadTable(reader io.Reader) (Table, error) {
  var table = make(Table)
  var byteData, err = ReadLongstr(reader)
  if err != nil {
    return nil, errors.New("Error reading table longstr: " + err.Error())
  }
  var data = bytes.NewBuffer(byteData)
  for data.Len() > 0 {
    key, err := ReadShortstr(data)
    if err != nil {
      return nil, errors.New("Error reading key: " + err.Error())
    }
    value, err := readValue(data)
    if err != nil {
      return nil, errors.New("Error reading value for " + key + ": " + err.Error())
    }
    table[key] = value
  }
  return table, nil
}

func readValue(reader io.Reader) (interface{}, error) {
  var t, err = ReadOctet(reader)
  if err != nil {
    return nil, err
  }

  switch {
  case t == 't':
    var v, err = ReadOctet(reader)
    return v != 0, err
  case t == 'b':
    var v int8
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'B':
    var v uint8
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'U':
    var v int16
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'u':
    var v uint16
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'I':
    var v int32
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'i':
    var v uint32
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'L':
    var v int64
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'l':
    var v uint64
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'f':
    var v float32
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'd':
    var v float64
    return v, binary.Read(reader, binary.BigEndian, &v)
  case t == 'D':
    var v = Decimal{}
    if err = binary.Read(reader, binary.BigEndian, &v.scale); err != nil {
      return nil, err
    }
    return v, binary.Read(reader, binary.BigEndian, &v.value)
  case t == 's':
    return ReadShortstr(reader)
  case t == 'S':
    return ReadLongstr(reader)
  case t == 'A':
    return readArray(reader)
  case t == 'T':
    return ReadTimestamp(reader)
  case t == 'F':
    return ReadTable(reader)
  case t == 'V':
    return nil, nil
  }
  return nil, errors.New("Unknown table value type")
}

func readArray(reader io.Reader) ([]interface{}, error) {
  var ret = make([]interface{}, 0, 1)
  var longstr, errs = ReadLongstr(reader)
  if errs != nil {
    return nil, errs
  }
  var data = bytes.NewBuffer(longstr)
  for data.Len() > 0 {
    var value, err = readValue(reader)
    if err != nil {
      return nil, err
    }
    ret = append(ret, value)
  }
  return ret, nil
}