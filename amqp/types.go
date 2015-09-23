package amqp

import (
  _ "fmt"
  "encoding/binary"
  "io"
  "errors"
  "bytes"
)

type Decimal struct {
  scale byte
  value int32
}

type Table map[string]interface{}

type MethodFrame interface {
  MethodIdentifier() (uint16, uint16)
  Read(reader io.Reader) (err error)
  Write(writer io.Writer) (err error)
}

type FrameWrapper struct {
  FrameType byte
  Channel uint16
  Payload []byte
}

func ReadTable(reader io.Reader) (Table, error) {
  var table = make(Table)
  var byteData, err = ReadLongstr(reader)
  if err != nil {
    return nil, err
  }
  var data = bytes.NewBuffer(byteData)
  for data.Len() > 0 {
    var key, errs = ReadShortstr(data)
    if errs != nil {
      return nil, errs
    }
    var value, errv = readValue(data)
    if errv != nil {
      return nil, errv
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