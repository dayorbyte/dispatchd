package amqp

import (
  //"encoding/binary"
  "io"
  "errors"
  //"bytes"
)



// **********************************************************************
//
//
//                    Connection
//
//
// **********************************************************************

var ClassIdConnection uint16 = 10

// **********************************************************************
//                    Connection - Start
// **********************************************************************

var MethodIdConnectionStart uint16 = 10
type ConnectionStart struct {
  VersionMajor byte
  VersionMinor byte
  ServerProperties Table
  Mechanisms []byte
  Locales []byte
}
func (f *ConnectionStart) Read(reader io.Reader) (err error) {
  f.VersionMajor, err = ReadOctet(reader)
  if err != nil {
    return errors.New("Error reading field VersionMajor")
  }

  f.VersionMinor, err = ReadOctet(reader)
  if err != nil {
    return errors.New("Error reading field VersionMinor")
  }

  f.ServerProperties, err = ReadPeerProperties(reader)
  if err != nil {
    return errors.New("Error reading field ServerProperties")
  }

  f.Mechanisms, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Mechanisms")
  }

  f.Locales, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Locales")
  }

  return
}
func (f *ConnectionStart) Write(writer io.Writer) (err error) {
 err = WriteOctet(writer, f.VersionMajor)
  if err != nil {
    return errors.New("Error writing field VersionMajor")
  }

 err = WriteOctet(writer, f.VersionMinor)
  if err != nil {
    return errors.New("Error writing field VersionMinor")
  }

 err = WritePeerProperties(writer, f.ServerProperties)
  if err != nil {
    return errors.New("Error writing field ServerProperties")
  }

 err = WriteLongstr(writer, f.Mechanisms)
  if err != nil {
    return errors.New("Error writing field Mechanisms")
  }

 err = WriteLongstr(writer, f.Locales)
  if err != nil {
    return errors.New("Error writing field Locales")
  }

  return
}

// **********************************************************************
//                    Connection - StartOk
// **********************************************************************

var MethodIdConnectionStartOk uint16 = 11
type ConnectionStartOk struct {
  ClientProperties Table
  Mechanism string
  Response []byte
  Locale string
}
func (f *ConnectionStartOk) Read(reader io.Reader) (err error) {
  f.ClientProperties, err = ReadPeerProperties(reader)
  if err != nil {
    return errors.New("Error reading field ClientProperties")
  }

  f.Mechanism, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Mechanism")
  }

  f.Response, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Response")
  }

  f.Locale, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Locale")
  }

  return
}
func (f *ConnectionStartOk) Write(writer io.Writer) (err error) {
 err = WritePeerProperties(writer, f.ClientProperties)
  if err != nil {
    return errors.New("Error writing field ClientProperties")
  }

 err = WriteShortstr(writer, f.Mechanism)
  if err != nil {
    return errors.New("Error writing field Mechanism")
  }

 err = WriteLongstr(writer, f.Response)
  if err != nil {
    return errors.New("Error writing field Response")
  }

 err = WriteShortstr(writer, f.Locale)
  if err != nil {
    return errors.New("Error writing field Locale")
  }

  return
}

// **********************************************************************
//                    Connection - Secure
// **********************************************************************

var MethodIdConnectionSecure uint16 = 20
type ConnectionSecure struct {
  Challenge []byte
}
func (f *ConnectionSecure) Read(reader io.Reader) (err error) {
  f.Challenge, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Challenge")
  }

  return
}
func (f *ConnectionSecure) Write(writer io.Writer) (err error) {
 err = WriteLongstr(writer, f.Challenge)
  if err != nil {
    return errors.New("Error writing field Challenge")
  }

  return
}

// **********************************************************************
//                    Connection - SecureOk
// **********************************************************************

var MethodIdConnectionSecureOk uint16 = 21
type ConnectionSecureOk struct {
  Response []byte
}
func (f *ConnectionSecureOk) Read(reader io.Reader) (err error) {
  f.Response, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Response")
  }

  return
}
func (f *ConnectionSecureOk) Write(writer io.Writer) (err error) {
 err = WriteLongstr(writer, f.Response)
  if err != nil {
    return errors.New("Error writing field Response")
  }

  return
}

// **********************************************************************
//                    Connection - Tune
// **********************************************************************

var MethodIdConnectionTune uint16 = 30
type ConnectionTune struct {
  ChannelMax uint16
  FrameMax uint32
  Heartbeat uint16
}
func (f *ConnectionTune) Read(reader io.Reader) (err error) {
  f.ChannelMax, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field ChannelMax")
  }

  f.FrameMax, err = ReadLong(reader)
  if err != nil {
    return errors.New("Error reading field FrameMax")
  }

  f.Heartbeat, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Heartbeat")
  }

  return
}
func (f *ConnectionTune) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.ChannelMax)
  if err != nil {
    return errors.New("Error writing field ChannelMax")
  }

 err = WriteLong(writer, f.FrameMax)
  if err != nil {
    return errors.New("Error writing field FrameMax")
  }

 err = WriteShort(writer, f.Heartbeat)
  if err != nil {
    return errors.New("Error writing field Heartbeat")
  }

  return
}

// **********************************************************************
//                    Connection - TuneOk
// **********************************************************************

var MethodIdConnectionTuneOk uint16 = 31
type ConnectionTuneOk struct {
  ChannelMax uint16
  FrameMax uint32
  Heartbeat uint16
}
func (f *ConnectionTuneOk) Read(reader io.Reader) (err error) {
  f.ChannelMax, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field ChannelMax")
  }

  f.FrameMax, err = ReadLong(reader)
  if err != nil {
    return errors.New("Error reading field FrameMax")
  }

  f.Heartbeat, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Heartbeat")
  }

  return
}
func (f *ConnectionTuneOk) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.ChannelMax)
  if err != nil {
    return errors.New("Error writing field ChannelMax")
  }

 err = WriteLong(writer, f.FrameMax)
  if err != nil {
    return errors.New("Error writing field FrameMax")
  }

 err = WriteShort(writer, f.Heartbeat)
  if err != nil {
    return errors.New("Error writing field Heartbeat")
  }

  return
}

// **********************************************************************
//                    Connection - Open
// **********************************************************************

var MethodIdConnectionOpen uint16 = 40
type ConnectionOpen struct {
  VirtualHost string
  Reserved1 string
  Reserved2 bool
}
func (f *ConnectionOpen) Read(reader io.Reader) (err error) {
  f.VirtualHost, err = ReadPath(reader)
  if err != nil {
    return errors.New("Error reading field VirtualHost")
  }

  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Reserved2, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Reserved2")
  }

  return
}
func (f *ConnectionOpen) Write(writer io.Writer) (err error) {
 err = WritePath(writer, f.VirtualHost)
  if err != nil {
    return errors.New("Error writing field VirtualHost")
  }

 err = WriteShortstr(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteBit(writer, f.Reserved2)
  if err != nil {
    return errors.New("Error writing field Reserved2")
  }

  return
}

// **********************************************************************
//                    Connection - OpenOk
// **********************************************************************

var MethodIdConnectionOpenOk uint16 = 41
type ConnectionOpenOk struct {
  Reserved1 string
}
func (f *ConnectionOpenOk) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *ConnectionOpenOk) Write(writer io.Writer) (err error) {
 err = WriteShortstr(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  return
}

// **********************************************************************
//                    Connection - Close
// **********************************************************************

var MethodIdConnectionClose uint16 = 50
type ConnectionClose struct {
  ReplyCode uint16
  ReplyText string
  ClassId uint16
  MethodId uint16
}
func (f *ConnectionClose) Read(reader io.Reader) (err error) {
  f.ReplyCode, err = ReadReplyCode(reader)
  if err != nil {
    return errors.New("Error reading field ReplyCode")
  }

  f.ReplyText, err = ReadReplyText(reader)
  if err != nil {
    return errors.New("Error reading field ReplyText")
  }

  f.ClassId, err = ReadClassId(reader)
  if err != nil {
    return errors.New("Error reading field ClassId")
  }

  f.MethodId, err = ReadMethodId(reader)
  if err != nil {
    return errors.New("Error reading field MethodId")
  }

  return
}
func (f *ConnectionClose) Write(writer io.Writer) (err error) {
 err = WriteReplyCode(writer, f.ReplyCode)
  if err != nil {
    return errors.New("Error writing field ReplyCode")
  }

 err = WriteReplyText(writer, f.ReplyText)
  if err != nil {
    return errors.New("Error writing field ReplyText")
  }

 err = WriteClassId(writer, f.ClassId)
  if err != nil {
    return errors.New("Error writing field ClassId")
  }

 err = WriteMethodId(writer, f.MethodId)
  if err != nil {
    return errors.New("Error writing field MethodId")
  }

  return
}

// **********************************************************************
//                    Connection - CloseOk
// **********************************************************************

var MethodIdConnectionCloseOk uint16 = 51
type ConnectionCloseOk struct {
}
func (f *ConnectionCloseOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ConnectionCloseOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//
//
//                    Channel
//
//
// **********************************************************************

var ClassIdChannel uint16 = 20

// **********************************************************************
//                    Channel - Open
// **********************************************************************

var MethodIdChannelOpen uint16 = 10
type ChannelOpen struct {
  Reserved1 string
}
func (f *ChannelOpen) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *ChannelOpen) Write(writer io.Writer) (err error) {
 err = WriteShortstr(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  return
}

// **********************************************************************
//                    Channel - OpenOk
// **********************************************************************

var MethodIdChannelOpenOk uint16 = 11
type ChannelOpenOk struct {
  Reserved1 []byte
}
func (f *ChannelOpenOk) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *ChannelOpenOk) Write(writer io.Writer) (err error) {
 err = WriteLongstr(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  return
}

// **********************************************************************
//                    Channel - Flow
// **********************************************************************

var MethodIdChannelFlow uint16 = 20
type ChannelFlow struct {
  Active bool
}
func (f *ChannelFlow) Read(reader io.Reader) (err error) {
  f.Active, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Active")
  }

  return
}
func (f *ChannelFlow) Write(writer io.Writer) (err error) {
 err = WriteBit(writer, f.Active)
  if err != nil {
    return errors.New("Error writing field Active")
  }

  return
}

// **********************************************************************
//                    Channel - FlowOk
// **********************************************************************

var MethodIdChannelFlowOk uint16 = 21
type ChannelFlowOk struct {
  Active bool
}
func (f *ChannelFlowOk) Read(reader io.Reader) (err error) {
  f.Active, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Active")
  }

  return
}
func (f *ChannelFlowOk) Write(writer io.Writer) (err error) {
 err = WriteBit(writer, f.Active)
  if err != nil {
    return errors.New("Error writing field Active")
  }

  return
}

// **********************************************************************
//                    Channel - Close
// **********************************************************************

var MethodIdChannelClose uint16 = 40
type ChannelClose struct {
  ReplyCode uint16
  ReplyText string
  ClassId uint16
  MethodId uint16
}
func (f *ChannelClose) Read(reader io.Reader) (err error) {
  f.ReplyCode, err = ReadReplyCode(reader)
  if err != nil {
    return errors.New("Error reading field ReplyCode")
  }

  f.ReplyText, err = ReadReplyText(reader)
  if err != nil {
    return errors.New("Error reading field ReplyText")
  }

  f.ClassId, err = ReadClassId(reader)
  if err != nil {
    return errors.New("Error reading field ClassId")
  }

  f.MethodId, err = ReadMethodId(reader)
  if err != nil {
    return errors.New("Error reading field MethodId")
  }

  return
}
func (f *ChannelClose) Write(writer io.Writer) (err error) {
 err = WriteReplyCode(writer, f.ReplyCode)
  if err != nil {
    return errors.New("Error writing field ReplyCode")
  }

 err = WriteReplyText(writer, f.ReplyText)
  if err != nil {
    return errors.New("Error writing field ReplyText")
  }

 err = WriteClassId(writer, f.ClassId)
  if err != nil {
    return errors.New("Error writing field ClassId")
  }

 err = WriteMethodId(writer, f.MethodId)
  if err != nil {
    return errors.New("Error writing field MethodId")
  }

  return
}

// **********************************************************************
//                    Channel - CloseOk
// **********************************************************************

var MethodIdChannelCloseOk uint16 = 41
type ChannelCloseOk struct {
}
func (f *ChannelCloseOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ChannelCloseOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//
//
//                    Exchange
//
//
// **********************************************************************

var ClassIdExchange uint16 = 40

// **********************************************************************
//                    Exchange - Declare
// **********************************************************************

var MethodIdExchangeDeclare uint16 = 10
type ExchangeDeclare struct {
  Reserved1 uint16
  Exchange string
  Type string
  Passive bool
  Durable bool
  Reserved2 bool
  Reserved3 bool
  NoWait bool
  Arguments Table
}
func (f *ExchangeDeclare) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.Type, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Type")
  }

  f.Passive, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Passive")
  }

  f.Durable, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Durable")
  }

  f.Reserved2, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Reserved2")
  }

  f.Reserved3, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Reserved3")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *ExchangeDeclare) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.Type)
  if err != nil {
    return errors.New("Error writing field Type")
  }

 err = WriteBit(writer, f.Passive)
  if err != nil {
    return errors.New("Error writing field Passive")
  }

 err = WriteBit(writer, f.Durable)
  if err != nil {
    return errors.New("Error writing field Durable")
  }

 err = WriteBit(writer, f.Reserved2)
  if err != nil {
    return errors.New("Error writing field Reserved2")
  }

 err = WriteBit(writer, f.Reserved3)
  if err != nil {
    return errors.New("Error writing field Reserved3")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

 err = WriteTable(writer, f.Arguments)
  if err != nil {
    return errors.New("Error writing field Arguments")
  }

  return
}

// **********************************************************************
//                    Exchange - DeclareOk
// **********************************************************************

var MethodIdExchangeDeclareOk uint16 = 11
type ExchangeDeclareOk struct {
}
func (f *ExchangeDeclareOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ExchangeDeclareOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Exchange - Delete
// **********************************************************************

var MethodIdExchangeDelete uint16 = 20
type ExchangeDelete struct {
  Reserved1 uint16
  Exchange string
  IfUnused bool
  NoWait bool
}
func (f *ExchangeDelete) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.IfUnused, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field IfUnused")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  return
}
func (f *ExchangeDelete) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteBit(writer, f.IfUnused)
  if err != nil {
    return errors.New("Error writing field IfUnused")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

  return
}

// **********************************************************************
//                    Exchange - DeleteOk
// **********************************************************************

var MethodIdExchangeDeleteOk uint16 = 21
type ExchangeDeleteOk struct {
}
func (f *ExchangeDeleteOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ExchangeDeleteOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//
//
//                    Queue
//
//
// **********************************************************************

var ClassIdQueue uint16 = 50

// **********************************************************************
//                    Queue - Declare
// **********************************************************************

var MethodIdQueueDeclare uint16 = 10
type QueueDeclare struct {
  Reserved1 uint16
  Queue string
  Passive bool
  Durable bool
  Exclusive bool
  AutoDelete bool
  NoWait bool
  Arguments Table
}
func (f *QueueDeclare) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.Passive, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Passive")
  }

  f.Durable, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Durable")
  }

  f.Exclusive, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Exclusive")
  }

  f.AutoDelete, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field AutoDelete")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *QueueDeclare) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteBit(writer, f.Passive)
  if err != nil {
    return errors.New("Error writing field Passive")
  }

 err = WriteBit(writer, f.Durable)
  if err != nil {
    return errors.New("Error writing field Durable")
  }

 err = WriteBit(writer, f.Exclusive)
  if err != nil {
    return errors.New("Error writing field Exclusive")
  }

 err = WriteBit(writer, f.AutoDelete)
  if err != nil {
    return errors.New("Error writing field AutoDelete")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

 err = WriteTable(writer, f.Arguments)
  if err != nil {
    return errors.New("Error writing field Arguments")
  }

  return
}

// **********************************************************************
//                    Queue - DeclareOk
// **********************************************************************

var MethodIdQueueDeclareOk uint16 = 11
type QueueDeclareOk struct {
  Queue string
  MessageCount uint32
  ConsumerCount uint32
}
func (f *QueueDeclareOk) Read(reader io.Reader) (err error) {
  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil {
    return errors.New("Error reading field MessageCount")
  }

  f.ConsumerCount, err = ReadLong(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerCount")
  }

  return
}
func (f *QueueDeclareOk) Write(writer io.Writer) (err error) {
 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteMessageCount(writer, f.MessageCount)
  if err != nil {
    return errors.New("Error writing field MessageCount")
  }

 err = WriteLong(writer, f.ConsumerCount)
  if err != nil {
    return errors.New("Error writing field ConsumerCount")
  }

  return
}

// **********************************************************************
//                    Queue - Bind
// **********************************************************************

var MethodIdQueueBind uint16 = 20
type QueueBind struct {
  Reserved1 uint16
  Queue string
  Exchange string
  RoutingKey string
  NoWait bool
  Arguments Table
}
func (f *QueueBind) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field RoutingKey")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *QueueBind) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.RoutingKey)
  if err != nil {
    return errors.New("Error writing field RoutingKey")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

 err = WriteTable(writer, f.Arguments)
  if err != nil {
    return errors.New("Error writing field Arguments")
  }

  return
}

// **********************************************************************
//                    Queue - BindOk
// **********************************************************************

var MethodIdQueueBindOk uint16 = 21
type QueueBindOk struct {
}
func (f *QueueBindOk) Read(reader io.Reader) (err error) {
  return
}
func (f *QueueBindOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Queue - Unbind
// **********************************************************************

var MethodIdQueueUnbind uint16 = 50
type QueueUnbind struct {
  Reserved1 uint16
  Queue string
  Exchange string
  RoutingKey string
  Arguments Table
}
func (f *QueueUnbind) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field RoutingKey")
  }

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *QueueUnbind) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.RoutingKey)
  if err != nil {
    return errors.New("Error writing field RoutingKey")
  }

 err = WriteTable(writer, f.Arguments)
  if err != nil {
    return errors.New("Error writing field Arguments")
  }

  return
}

// **********************************************************************
//                    Queue - UnbindOk
// **********************************************************************

var MethodIdQueueUnbindOk uint16 = 51
type QueueUnbindOk struct {
}
func (f *QueueUnbindOk) Read(reader io.Reader) (err error) {
  return
}
func (f *QueueUnbindOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Queue - Purge
// **********************************************************************

var MethodIdQueuePurge uint16 = 30
type QueuePurge struct {
  Reserved1 uint16
  Queue string
  NoWait bool
}
func (f *QueuePurge) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  return
}
func (f *QueuePurge) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

  return
}

// **********************************************************************
//                    Queue - PurgeOk
// **********************************************************************

var MethodIdQueuePurgeOk uint16 = 31
type QueuePurgeOk struct {
  MessageCount uint32
}
func (f *QueuePurgeOk) Read(reader io.Reader) (err error) {
  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil {
    return errors.New("Error reading field MessageCount")
  }

  return
}
func (f *QueuePurgeOk) Write(writer io.Writer) (err error) {
 err = WriteMessageCount(writer, f.MessageCount)
  if err != nil {
    return errors.New("Error writing field MessageCount")
  }

  return
}

// **********************************************************************
//                    Queue - Delete
// **********************************************************************

var MethodIdQueueDelete uint16 = 40
type QueueDelete struct {
  Reserved1 uint16
  Queue string
  IfUnused bool
  IfEmpty bool
  NoWait bool
}
func (f *QueueDelete) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.IfUnused, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field IfUnused")
  }

  f.IfEmpty, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field IfEmpty")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  return
}
func (f *QueueDelete) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteBit(writer, f.IfUnused)
  if err != nil {
    return errors.New("Error writing field IfUnused")
  }

 err = WriteBit(writer, f.IfEmpty)
  if err != nil {
    return errors.New("Error writing field IfEmpty")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

  return
}

// **********************************************************************
//                    Queue - DeleteOk
// **********************************************************************

var MethodIdQueueDeleteOk uint16 = 41
type QueueDeleteOk struct {
  MessageCount uint32
}
func (f *QueueDeleteOk) Read(reader io.Reader) (err error) {
  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil {
    return errors.New("Error reading field MessageCount")
  }

  return
}
func (f *QueueDeleteOk) Write(writer io.Writer) (err error) {
 err = WriteMessageCount(writer, f.MessageCount)
  if err != nil {
    return errors.New("Error writing field MessageCount")
  }

  return
}

// **********************************************************************
//
//
//                    Basic
//
//
// **********************************************************************

var ClassIdBasic uint16 = 60

// **********************************************************************
//                    Basic - Qos
// **********************************************************************

var MethodIdBasicQos uint16 = 10
type BasicQos struct {
  PrefetchSize uint32
  PrefetchCount uint16
  Global bool
}
func (f *BasicQos) Read(reader io.Reader) (err error) {
  f.PrefetchSize, err = ReadLong(reader)
  if err != nil {
    return errors.New("Error reading field PrefetchSize")
  }

  f.PrefetchCount, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field PrefetchCount")
  }

  f.Global, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Global")
  }

  return
}
func (f *BasicQos) Write(writer io.Writer) (err error) {
 err = WriteLong(writer, f.PrefetchSize)
  if err != nil {
    return errors.New("Error writing field PrefetchSize")
  }

 err = WriteShort(writer, f.PrefetchCount)
  if err != nil {
    return errors.New("Error writing field PrefetchCount")
  }

 err = WriteBit(writer, f.Global)
  if err != nil {
    return errors.New("Error writing field Global")
  }

  return
}

// **********************************************************************
//                    Basic - QosOk
// **********************************************************************

var MethodIdBasicQosOk uint16 = 11
type BasicQosOk struct {
}
func (f *BasicQosOk) Read(reader io.Reader) (err error) {
  return
}
func (f *BasicQosOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Basic - Consume
// **********************************************************************

var MethodIdBasicConsume uint16 = 20
type BasicConsume struct {
  Reserved1 uint16
  Queue string
  ConsumerTag string
  NoLocal bool
  NoAck bool
  Exclusive bool
  NoWait bool
  Arguments Table
}
func (f *BasicConsume) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  f.NoLocal, err = ReadNoLocal(reader)
  if err != nil {
    return errors.New("Error reading field NoLocal")
  }

  f.NoAck, err = ReadNoAck(reader)
  if err != nil {
    return errors.New("Error reading field NoAck")
  }

  f.Exclusive, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Exclusive")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *BasicConsume) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

 err = WriteNoLocal(writer, f.NoLocal)
  if err != nil {
    return errors.New("Error writing field NoLocal")
  }

 err = WriteNoAck(writer, f.NoAck)
  if err != nil {
    return errors.New("Error writing field NoAck")
  }

 err = WriteBit(writer, f.Exclusive)
  if err != nil {
    return errors.New("Error writing field Exclusive")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

 err = WriteTable(writer, f.Arguments)
  if err != nil {
    return errors.New("Error writing field Arguments")
  }

  return
}

// **********************************************************************
//                    Basic - ConsumeOk
// **********************************************************************

var MethodIdBasicConsumeOk uint16 = 21
type BasicConsumeOk struct {
  ConsumerTag string
}
func (f *BasicConsumeOk) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  return
}
func (f *BasicConsumeOk) Write(writer io.Writer) (err error) {
 err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

  return
}

// **********************************************************************
//                    Basic - Cancel
// **********************************************************************

var MethodIdBasicCancel uint16 = 30
type BasicCancel struct {
  ConsumerTag string
  NoWait bool
}
func (f *BasicCancel) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  f.NoWait, err = ReadNoWait(reader)
  if err != nil {
    return errors.New("Error reading field NoWait")
  }

  return
}
func (f *BasicCancel) Write(writer io.Writer) (err error) {
 err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

 err = WriteNoWait(writer, f.NoWait)
  if err != nil {
    return errors.New("Error writing field NoWait")
  }

  return
}

// **********************************************************************
//                    Basic - CancelOk
// **********************************************************************

var MethodIdBasicCancelOk uint16 = 31
type BasicCancelOk struct {
  ConsumerTag string
}
func (f *BasicCancelOk) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  return
}
func (f *BasicCancelOk) Write(writer io.Writer) (err error) {
 err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

  return
}

// **********************************************************************
//                    Basic - Publish
// **********************************************************************

var MethodIdBasicPublish uint16 = 40
type BasicPublish struct {
  Reserved1 uint16
  Exchange string
  RoutingKey string
  Mandatory bool
  Immediate bool
}
func (f *BasicPublish) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field RoutingKey")
  }

  f.Mandatory, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Mandatory")
  }

  f.Immediate, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Immediate")
  }

  return
}
func (f *BasicPublish) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.RoutingKey)
  if err != nil {
    return errors.New("Error writing field RoutingKey")
  }

 err = WriteBit(writer, f.Mandatory)
  if err != nil {
    return errors.New("Error writing field Mandatory")
  }

 err = WriteBit(writer, f.Immediate)
  if err != nil {
    return errors.New("Error writing field Immediate")
  }

  return
}

// **********************************************************************
//                    Basic - Return
// **********************************************************************

var MethodIdBasicReturn uint16 = 50
type BasicReturn struct {
  ReplyCode uint16
  ReplyText string
  Exchange string
  RoutingKey string
}
func (f *BasicReturn) Read(reader io.Reader) (err error) {
  f.ReplyCode, err = ReadReplyCode(reader)
  if err != nil {
    return errors.New("Error reading field ReplyCode")
  }

  f.ReplyText, err = ReadReplyText(reader)
  if err != nil {
    return errors.New("Error reading field ReplyText")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field RoutingKey")
  }

  return
}
func (f *BasicReturn) Write(writer io.Writer) (err error) {
 err = WriteReplyCode(writer, f.ReplyCode)
  if err != nil {
    return errors.New("Error writing field ReplyCode")
  }

 err = WriteReplyText(writer, f.ReplyText)
  if err != nil {
    return errors.New("Error writing field ReplyText")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.RoutingKey)
  if err != nil {
    return errors.New("Error writing field RoutingKey")
  }

  return
}

// **********************************************************************
//                    Basic - Deliver
// **********************************************************************

var MethodIdBasicDeliver uint16 = 60
type BasicDeliver struct {
  ConsumerTag string
  DeliveryTag uint64
  Redelivered bool
  Exchange string
  RoutingKey string
}
func (f *BasicDeliver) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }

  f.Redelivered, err = ReadRedelivered(reader)
  if err != nil {
    return errors.New("Error reading field Redelivered")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field RoutingKey")
  }

  return
}
func (f *BasicDeliver) Write(writer io.Writer) (err error) {
 err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

 err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

 err = WriteRedelivered(writer, f.Redelivered)
  if err != nil {
    return errors.New("Error writing field Redelivered")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.RoutingKey)
  if err != nil {
    return errors.New("Error writing field RoutingKey")
  }

  return
}

// **********************************************************************
//                    Basic - Get
// **********************************************************************

var MethodIdBasicGet uint16 = 70
type BasicGet struct {
  Reserved1 uint16
  Queue string
  NoAck bool
}
func (f *BasicGet) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShort(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  f.Queue, err = ReadQueueName(reader)
  if err != nil {
    return errors.New("Error reading field Queue")
  }

  f.NoAck, err = ReadNoAck(reader)
  if err != nil {
    return errors.New("Error reading field NoAck")
  }

  return
}
func (f *BasicGet) Write(writer io.Writer) (err error) {
 err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

 err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

 err = WriteNoAck(writer, f.NoAck)
  if err != nil {
    return errors.New("Error writing field NoAck")
  }

  return
}

// **********************************************************************
//                    Basic - GetOk
// **********************************************************************

var MethodIdBasicGetOk uint16 = 71
type BasicGetOk struct {
  DeliveryTag uint64
  Redelivered bool
  Exchange string
  RoutingKey string
  MessageCount uint32
}
func (f *BasicGetOk) Read(reader io.Reader) (err error) {
  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }

  f.Redelivered, err = ReadRedelivered(reader)
  if err != nil {
    return errors.New("Error reading field Redelivered")
  }

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil {
    return errors.New("Error reading field Exchange")
  }

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field RoutingKey")
  }

  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil {
    return errors.New("Error reading field MessageCount")
  }

  return
}
func (f *BasicGetOk) Write(writer io.Writer) (err error) {
 err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

 err = WriteRedelivered(writer, f.Redelivered)
  if err != nil {
    return errors.New("Error writing field Redelivered")
  }

 err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

 err = WriteShortstr(writer, f.RoutingKey)
  if err != nil {
    return errors.New("Error writing field RoutingKey")
  }

 err = WriteMessageCount(writer, f.MessageCount)
  if err != nil {
    return errors.New("Error writing field MessageCount")
  }

  return
}

// **********************************************************************
//                    Basic - GetEmpty
// **********************************************************************

var MethodIdBasicGetEmpty uint16 = 72
type BasicGetEmpty struct {
  Reserved1 string
}
func (f *BasicGetEmpty) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *BasicGetEmpty) Write(writer io.Writer) (err error) {
 err = WriteShortstr(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  return
}

// **********************************************************************
//                    Basic - Ack
// **********************************************************************

var MethodIdBasicAck uint16 = 80
type BasicAck struct {
  DeliveryTag uint64
  Multiple bool
}
func (f *BasicAck) Read(reader io.Reader) (err error) {
  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }

  f.Multiple, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Multiple")
  }

  return
}
func (f *BasicAck) Write(writer io.Writer) (err error) {
 err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

 err = WriteBit(writer, f.Multiple)
  if err != nil {
    return errors.New("Error writing field Multiple")
  }

  return
}

// **********************************************************************
//                    Basic - Reject
// **********************************************************************

var MethodIdBasicReject uint16 = 90
type BasicReject struct {
  DeliveryTag uint64
  Requeue bool
}
func (f *BasicReject) Read(reader io.Reader) (err error) {
  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }

  f.Requeue, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Requeue")
  }

  return
}
func (f *BasicReject) Write(writer io.Writer) (err error) {
 err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

 err = WriteBit(writer, f.Requeue)
  if err != nil {
    return errors.New("Error writing field Requeue")
  }

  return
}

// **********************************************************************
//                    Basic - RecoverAsync
// **********************************************************************

var MethodIdBasicRecoverAsync uint16 = 100
type BasicRecoverAsync struct {
  Requeue bool
}
func (f *BasicRecoverAsync) Read(reader io.Reader) (err error) {
  f.Requeue, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Requeue")
  }

  return
}
func (f *BasicRecoverAsync) Write(writer io.Writer) (err error) {
 err = WriteBit(writer, f.Requeue)
  if err != nil {
    return errors.New("Error writing field Requeue")
  }

  return
}

// **********************************************************************
//                    Basic - Recover
// **********************************************************************

var MethodIdBasicRecover uint16 = 110
type BasicRecover struct {
  Requeue bool
}
func (f *BasicRecover) Read(reader io.Reader) (err error) {
  f.Requeue, err = ReadBit(reader)
  if err != nil {
    return errors.New("Error reading field Requeue")
  }

  return
}
func (f *BasicRecover) Write(writer io.Writer) (err error) {
 err = WriteBit(writer, f.Requeue)
  if err != nil {
    return errors.New("Error writing field Requeue")
  }

  return
}

// **********************************************************************
//                    Basic - RecoverOk
// **********************************************************************

var MethodIdBasicRecoverOk uint16 = 111
type BasicRecoverOk struct {
}
func (f *BasicRecoverOk) Read(reader io.Reader) (err error) {
  return
}
func (f *BasicRecoverOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//
//
//                    Tx
//
//
// **********************************************************************

var ClassIdTx uint16 = 90

// **********************************************************************
//                    Tx - Select
// **********************************************************************

var MethodIdTxSelect uint16 = 10
type TxSelect struct {
}
func (f *TxSelect) Read(reader io.Reader) (err error) {
  return
}
func (f *TxSelect) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Tx - SelectOk
// **********************************************************************

var MethodIdTxSelectOk uint16 = 11
type TxSelectOk struct {
}
func (f *TxSelectOk) Read(reader io.Reader) (err error) {
  return
}
func (f *TxSelectOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Tx - Commit
// **********************************************************************

var MethodIdTxCommit uint16 = 20
type TxCommit struct {
}
func (f *TxCommit) Read(reader io.Reader) (err error) {
  return
}
func (f *TxCommit) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Tx - CommitOk
// **********************************************************************

var MethodIdTxCommitOk uint16 = 21
type TxCommitOk struct {
}
func (f *TxCommitOk) Read(reader io.Reader) (err error) {
  return
}
func (f *TxCommitOk) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Tx - Rollback
// **********************************************************************

var MethodIdTxRollback uint16 = 30
type TxRollback struct {
}
func (f *TxRollback) Read(reader io.Reader) (err error) {
  return
}
func (f *TxRollback) Write(writer io.Writer) (err error) {
  return
}

// **********************************************************************
//                    Tx - RollbackOk
// **********************************************************************

var MethodIdTxRollbackOk uint16 = 31
type TxRollbackOk struct {
}
func (f *TxRollbackOk) Read(reader io.Reader) (err error) {
  return
}
func (f *TxRollbackOk) Write(writer io.Writer) (err error) {
  return
}
