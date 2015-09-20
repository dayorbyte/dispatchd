package amqp

import (
  "encoding/binary"
  "io"
  "errors"
  "bytes"
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

var MethodIdStart uint16 = 10
type ConnectionStart struct {
  VersionMajor byte
  VersionMinor byte
  ServerProperties Table
  Mechanisms []byte
  Locales []byte
}
func (f *ConnectionStart) Read(reader io.Reader) {  f.VersionMajor, err = ReadOctet(reader)
  if err != nil:
    return errors.New("Error reading field VersionMajor")

  f.VersionMinor, err = ReadOctet(reader)
  if err != nil:
    return errors.New("Error reading field VersionMinor")

  f.ServerProperties, err = ReadPeerProperties(reader)
  if err != nil:
    return errors.New("Error reading field ServerProperties")

  f.Mechanisms, err = ReadLongstr(reader)
  if err != nil:
    return errors.New("Error reading field Mechanisms")

  f.Locales, err = ReadLongstr(reader)
  if err != nil:
    return errors.New("Error reading field Locales")

}
func (f *ConnectionStart) Write(writer io.Writer) {  err = WriteOctet(writer)
  if err != nil:
    return errors.New("Error writing field VersionMajor")

  err = WriteOctet(writer)
  if err != nil:
    return errors.New("Error writing field VersionMinor")

  err = WritePeerProperties(writer)
  if err != nil:
    return errors.New("Error writing field ServerProperties")

  err = WriteLongstr(writer)
  if err != nil:
    return errors.New("Error writing field Mechanisms")

  err = WriteLongstr(writer)
  if err != nil:
    return errors.New("Error writing field Locales")

}

// **********************************************************************
//                    Connection - StartOk
// **********************************************************************

var MethodIdStartOk uint16 = 11
type ConnectionStartOk struct {
  ClientProperties Table
  Mechanism string
  Response []byte
  Locale string
}
func (f *ConnectionStartOk) Read(reader io.Reader) {  f.ClientProperties, err = ReadPeerProperties(reader)
  if err != nil:
    return errors.New("Error reading field ClientProperties")

  f.Mechanism, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Mechanism")

  f.Response, err = ReadLongstr(reader)
  if err != nil:
    return errors.New("Error reading field Response")

  f.Locale, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Locale")

}
func (f *ConnectionStartOk) Write(writer io.Writer) {  err = WritePeerProperties(writer)
  if err != nil:
    return errors.New("Error writing field ClientProperties")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Mechanism")

  err = WriteLongstr(writer)
  if err != nil:
    return errors.New("Error writing field Response")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Locale")

}

// **********************************************************************
//                    Connection - Secure
// **********************************************************************

var MethodIdSecure uint16 = 20
type ConnectionSecure struct {
  Challenge []byte
}
func (f *ConnectionSecure) Read(reader io.Reader) {  f.Challenge, err = ReadLongstr(reader)
  if err != nil:
    return errors.New("Error reading field Challenge")

}
func (f *ConnectionSecure) Write(writer io.Writer) {  err = WriteLongstr(writer)
  if err != nil:
    return errors.New("Error writing field Challenge")

}

// **********************************************************************
//                    Connection - SecureOk
// **********************************************************************

var MethodIdSecureOk uint16 = 21
type ConnectionSecureOk struct {
  Response []byte
}
func (f *ConnectionSecureOk) Read(reader io.Reader) {  f.Response, err = ReadLongstr(reader)
  if err != nil:
    return errors.New("Error reading field Response")

}
func (f *ConnectionSecureOk) Write(writer io.Writer) {  err = WriteLongstr(writer)
  if err != nil:
    return errors.New("Error writing field Response")

}

// **********************************************************************
//                    Connection - Tune
// **********************************************************************

var MethodIdTune uint16 = 30
type ConnectionTune struct {
  ChannelMax uint16
  FrameMax uint32
  Heartbeat uint16
}
func (f *ConnectionTune) Read(reader io.Reader) {  f.ChannelMax, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field ChannelMax")

  f.FrameMax, err = ReadLong(reader)
  if err != nil:
    return errors.New("Error reading field FrameMax")

  f.Heartbeat, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Heartbeat")

}
func (f *ConnectionTune) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field ChannelMax")

  err = WriteLong(writer)
  if err != nil:
    return errors.New("Error writing field FrameMax")

  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Heartbeat")

}

// **********************************************************************
//                    Connection - TuneOk
// **********************************************************************

var MethodIdTuneOk uint16 = 31
type ConnectionTuneOk struct {
  ChannelMax uint16
  FrameMax uint32
  Heartbeat uint16
}
func (f *ConnectionTuneOk) Read(reader io.Reader) {  f.ChannelMax, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field ChannelMax")

  f.FrameMax, err = ReadLong(reader)
  if err != nil:
    return errors.New("Error reading field FrameMax")

  f.Heartbeat, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Heartbeat")

}
func (f *ConnectionTuneOk) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field ChannelMax")

  err = WriteLong(writer)
  if err != nil:
    return errors.New("Error writing field FrameMax")

  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Heartbeat")

}

// **********************************************************************
//                    Connection - Open
// **********************************************************************

var MethodIdOpen uint16 = 40
type ConnectionOpen struct {
  VirtualHost string
  Reserved1 string
  Reserved2 bool
}
func (f *ConnectionOpen) Read(reader io.Reader) {  f.VirtualHost, err = ReadPath(reader)
  if err != nil:
    return errors.New("Error reading field VirtualHost")

  f.Reserved1, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Reserved2, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Reserved2")

}
func (f *ConnectionOpen) Write(writer io.Writer) {  err = WritePath(writer)
  if err != nil:
    return errors.New("Error writing field VirtualHost")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Reserved2")

}

// **********************************************************************
//                    Connection - OpenOk
// **********************************************************************

var MethodIdOpenOk uint16 = 41
type ConnectionOpenOk struct {
  Reserved1 string
}
func (f *ConnectionOpenOk) Read(reader io.Reader) {  f.Reserved1, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

}
func (f *ConnectionOpenOk) Write(writer io.Writer) {  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

}

// **********************************************************************
//                    Connection - Close
// **********************************************************************

var MethodIdClose uint16 = 50
type ConnectionClose struct {
  ReplyCode uint16
  ReplyText string
  ClassId uint16
  MethodId uint16
}
func (f *ConnectionClose) Read(reader io.Reader) {  f.ReplyCode, err = ReadReplyCode(reader)
  if err != nil:
    return errors.New("Error reading field ReplyCode")

  f.ReplyText, err = ReadReplyText(reader)
  if err != nil:
    return errors.New("Error reading field ReplyText")

  f.ClassId, err = ReadClassId(reader)
  if err != nil:
    return errors.New("Error reading field ClassId")

  f.MethodId, err = ReadMethodId(reader)
  if err != nil:
    return errors.New("Error reading field MethodId")

}
func (f *ConnectionClose) Write(writer io.Writer) {  err = WriteReplyCode(writer)
  if err != nil:
    return errors.New("Error writing field ReplyCode")

  err = WriteReplyText(writer)
  if err != nil:
    return errors.New("Error writing field ReplyText")

  err = WriteClassId(writer)
  if err != nil:
    return errors.New("Error writing field ClassId")

  err = WriteMethodId(writer)
  if err != nil:
    return errors.New("Error writing field MethodId")

}

// **********************************************************************
//                    Connection - CloseOk
// **********************************************************************

var MethodIdCloseOk uint16 = 51
type ConnectionCloseOk struct {
}
func (f *ConnectionCloseOk) Read(reader io.Reader) {}
func (f *ConnectionCloseOk) Write(writer io.Writer) {}

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

var MethodIdOpen uint16 = 10
type ChannelOpen struct {
  Reserved1 string
}
func (f *ChannelOpen) Read(reader io.Reader) {  f.Reserved1, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

}
func (f *ChannelOpen) Write(writer io.Writer) {  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

}

// **********************************************************************
//                    Channel - OpenOk
// **********************************************************************

var MethodIdOpenOk uint16 = 11
type ChannelOpenOk struct {
  Reserved1 []byte
}
func (f *ChannelOpenOk) Read(reader io.Reader) {  f.Reserved1, err = ReadLongstr(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

}
func (f *ChannelOpenOk) Write(writer io.Writer) {  err = WriteLongstr(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

}

// **********************************************************************
//                    Channel - Flow
// **********************************************************************

var MethodIdFlow uint16 = 20
type ChannelFlow struct {
  Active bool
}
func (f *ChannelFlow) Read(reader io.Reader) {  f.Active, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Active")

}
func (f *ChannelFlow) Write(writer io.Writer) {  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Active")

}

// **********************************************************************
//                    Channel - FlowOk
// **********************************************************************

var MethodIdFlowOk uint16 = 21
type ChannelFlowOk struct {
  Active bool
}
func (f *ChannelFlowOk) Read(reader io.Reader) {  f.Active, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Active")

}
func (f *ChannelFlowOk) Write(writer io.Writer) {  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Active")

}

// **********************************************************************
//                    Channel - Close
// **********************************************************************

var MethodIdClose uint16 = 40
type ChannelClose struct {
  ReplyCode uint16
  ReplyText string
  ClassId uint16
  MethodId uint16
}
func (f *ChannelClose) Read(reader io.Reader) {  f.ReplyCode, err = ReadReplyCode(reader)
  if err != nil:
    return errors.New("Error reading field ReplyCode")

  f.ReplyText, err = ReadReplyText(reader)
  if err != nil:
    return errors.New("Error reading field ReplyText")

  f.ClassId, err = ReadClassId(reader)
  if err != nil:
    return errors.New("Error reading field ClassId")

  f.MethodId, err = ReadMethodId(reader)
  if err != nil:
    return errors.New("Error reading field MethodId")

}
func (f *ChannelClose) Write(writer io.Writer) {  err = WriteReplyCode(writer)
  if err != nil:
    return errors.New("Error writing field ReplyCode")

  err = WriteReplyText(writer)
  if err != nil:
    return errors.New("Error writing field ReplyText")

  err = WriteClassId(writer)
  if err != nil:
    return errors.New("Error writing field ClassId")

  err = WriteMethodId(writer)
  if err != nil:
    return errors.New("Error writing field MethodId")

}

// **********************************************************************
//                    Channel - CloseOk
// **********************************************************************

var MethodIdCloseOk uint16 = 41
type ChannelCloseOk struct {
}
func (f *ChannelCloseOk) Read(reader io.Reader) {}
func (f *ChannelCloseOk) Write(writer io.Writer) {}

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

var MethodIdDeclare uint16 = 10
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
func (f *ExchangeDeclare) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.Type, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Type")

  f.Passive, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Passive")

  f.Durable, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Durable")

  f.Reserved2, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Reserved2")

  f.Reserved3, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Reserved3")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

  f.Arguments, err = ReadTable(reader)
  if err != nil:
    return errors.New("Error reading field Arguments")

}
func (f *ExchangeDeclare) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Type")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Passive")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Durable")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Reserved2")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Reserved3")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

  err = WriteTable(writer)
  if err != nil:
    return errors.New("Error writing field Arguments")

}

// **********************************************************************
//                    Exchange - DeclareOk
// **********************************************************************

var MethodIdDeclareOk uint16 = 11
type ExchangeDeclareOk struct {
}
func (f *ExchangeDeclareOk) Read(reader io.Reader) {}
func (f *ExchangeDeclareOk) Write(writer io.Writer) {}

// **********************************************************************
//                    Exchange - Delete
// **********************************************************************

var MethodIdDelete uint16 = 20
type ExchangeDelete struct {
  Reserved1 uint16
  Exchange string
  IfUnused bool
  NoWait bool
}
func (f *ExchangeDelete) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.IfUnused, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field IfUnused")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

}
func (f *ExchangeDelete) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field IfUnused")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

}

// **********************************************************************
//                    Exchange - DeleteOk
// **********************************************************************

var MethodIdDeleteOk uint16 = 21
type ExchangeDeleteOk struct {
}
func (f *ExchangeDeleteOk) Read(reader io.Reader) {}
func (f *ExchangeDeleteOk) Write(writer io.Writer) {}

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

var MethodIdDeclare uint16 = 10
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
func (f *QueueDeclare) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.Passive, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Passive")

  f.Durable, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Durable")

  f.Exclusive, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Exclusive")

  f.AutoDelete, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field AutoDelete")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

  f.Arguments, err = ReadTable(reader)
  if err != nil:
    return errors.New("Error reading field Arguments")

}
func (f *QueueDeclare) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Passive")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Durable")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Exclusive")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field AutoDelete")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

  err = WriteTable(writer)
  if err != nil:
    return errors.New("Error writing field Arguments")

}

// **********************************************************************
//                    Queue - DeclareOk
// **********************************************************************

var MethodIdDeclareOk uint16 = 11
type QueueDeclareOk struct {
  Queue string
  MessageCount uint32
  ConsumerCount uint32
}
func (f *QueueDeclareOk) Read(reader io.Reader) {  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil:
    return errors.New("Error reading field MessageCount")

  f.ConsumerCount, err = ReadLong(reader)
  if err != nil:
    return errors.New("Error reading field ConsumerCount")

}
func (f *QueueDeclareOk) Write(writer io.Writer) {  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteMessageCount(writer)
  if err != nil:
    return errors.New("Error writing field MessageCount")

  err = WriteLong(writer)
  if err != nil:
    return errors.New("Error writing field ConsumerCount")

}

// **********************************************************************
//                    Queue - Bind
// **********************************************************************

var MethodIdBind uint16 = 20
type QueueBind struct {
  Reserved1 uint16
  Queue string
  Exchange string
  RoutingKey string
  NoWait bool
  Arguments Table
}
func (f *QueueBind) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field RoutingKey")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

  f.Arguments, err = ReadTable(reader)
  if err != nil:
    return errors.New("Error reading field Arguments")

}
func (f *QueueBind) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field RoutingKey")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

  err = WriteTable(writer)
  if err != nil:
    return errors.New("Error writing field Arguments")

}

// **********************************************************************
//                    Queue - BindOk
// **********************************************************************

var MethodIdBindOk uint16 = 21
type QueueBindOk struct {
}
func (f *QueueBindOk) Read(reader io.Reader) {}
func (f *QueueBindOk) Write(writer io.Writer) {}

// **********************************************************************
//                    Queue - Unbind
// **********************************************************************

var MethodIdUnbind uint16 = 50
type QueueUnbind struct {
  Reserved1 uint16
  Queue string
  Exchange string
  RoutingKey string
  Arguments Table
}
func (f *QueueUnbind) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field RoutingKey")

  f.Arguments, err = ReadTable(reader)
  if err != nil:
    return errors.New("Error reading field Arguments")

}
func (f *QueueUnbind) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field RoutingKey")

  err = WriteTable(writer)
  if err != nil:
    return errors.New("Error writing field Arguments")

}

// **********************************************************************
//                    Queue - UnbindOk
// **********************************************************************

var MethodIdUnbindOk uint16 = 51
type QueueUnbindOk struct {
}
func (f *QueueUnbindOk) Read(reader io.Reader) {}
func (f *QueueUnbindOk) Write(writer io.Writer) {}

// **********************************************************************
//                    Queue - Purge
// **********************************************************************

var MethodIdPurge uint16 = 30
type QueuePurge struct {
  Reserved1 uint16
  Queue string
  NoWait bool
}
func (f *QueuePurge) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

}
func (f *QueuePurge) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

}

// **********************************************************************
//                    Queue - PurgeOk
// **********************************************************************

var MethodIdPurgeOk uint16 = 31
type QueuePurgeOk struct {
  MessageCount uint32
}
func (f *QueuePurgeOk) Read(reader io.Reader) {  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil:
    return errors.New("Error reading field MessageCount")

}
func (f *QueuePurgeOk) Write(writer io.Writer) {  err = WriteMessageCount(writer)
  if err != nil:
    return errors.New("Error writing field MessageCount")

}

// **********************************************************************
//                    Queue - Delete
// **********************************************************************

var MethodIdDelete uint16 = 40
type QueueDelete struct {
  Reserved1 uint16
  Queue string
  IfUnused bool
  IfEmpty bool
  NoWait bool
}
func (f *QueueDelete) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.IfUnused, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field IfUnused")

  f.IfEmpty, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field IfEmpty")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

}
func (f *QueueDelete) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field IfUnused")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field IfEmpty")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

}

// **********************************************************************
//                    Queue - DeleteOk
// **********************************************************************

var MethodIdDeleteOk uint16 = 41
type QueueDeleteOk struct {
  MessageCount uint32
}
func (f *QueueDeleteOk) Read(reader io.Reader) {  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil:
    return errors.New("Error reading field MessageCount")

}
func (f *QueueDeleteOk) Write(writer io.Writer) {  err = WriteMessageCount(writer)
  if err != nil:
    return errors.New("Error writing field MessageCount")

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

var MethodIdQos uint16 = 10
type BasicQos struct {
  PrefetchSize uint32
  PrefetchCount uint16
  Global bool
}
func (f *BasicQos) Read(reader io.Reader) {  f.PrefetchSize, err = ReadLong(reader)
  if err != nil:
    return errors.New("Error reading field PrefetchSize")

  f.PrefetchCount, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field PrefetchCount")

  f.Global, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Global")

}
func (f *BasicQos) Write(writer io.Writer) {  err = WriteLong(writer)
  if err != nil:
    return errors.New("Error writing field PrefetchSize")

  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field PrefetchCount")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Global")

}

// **********************************************************************
//                    Basic - QosOk
// **********************************************************************

var MethodIdQosOk uint16 = 11
type BasicQosOk struct {
}
func (f *BasicQosOk) Read(reader io.Reader) {}
func (f *BasicQosOk) Write(writer io.Writer) {}

// **********************************************************************
//                    Basic - Consume
// **********************************************************************

var MethodIdConsume uint16 = 20
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
func (f *BasicConsume) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil:
    return errors.New("Error reading field ConsumerTag")

  f.NoLocal, err = ReadNoLocal(reader)
  if err != nil:
    return errors.New("Error reading field NoLocal")

  f.NoAck, err = ReadNoAck(reader)
  if err != nil:
    return errors.New("Error reading field NoAck")

  f.Exclusive, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Exclusive")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

  f.Arguments, err = ReadTable(reader)
  if err != nil:
    return errors.New("Error reading field Arguments")

}
func (f *BasicConsume) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteConsumerTag(writer)
  if err != nil:
    return errors.New("Error writing field ConsumerTag")

  err = WriteNoLocal(writer)
  if err != nil:
    return errors.New("Error writing field NoLocal")

  err = WriteNoAck(writer)
  if err != nil:
    return errors.New("Error writing field NoAck")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Exclusive")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

  err = WriteTable(writer)
  if err != nil:
    return errors.New("Error writing field Arguments")

}

// **********************************************************************
//                    Basic - ConsumeOk
// **********************************************************************

var MethodIdConsumeOk uint16 = 21
type BasicConsumeOk struct {
  ConsumerTag string
}
func (f *BasicConsumeOk) Read(reader io.Reader) {  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil:
    return errors.New("Error reading field ConsumerTag")

}
func (f *BasicConsumeOk) Write(writer io.Writer) {  err = WriteConsumerTag(writer)
  if err != nil:
    return errors.New("Error writing field ConsumerTag")

}

// **********************************************************************
//                    Basic - Cancel
// **********************************************************************

var MethodIdCancel uint16 = 30
type BasicCancel struct {
  ConsumerTag string
  NoWait bool
}
func (f *BasicCancel) Read(reader io.Reader) {  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil:
    return errors.New("Error reading field ConsumerTag")

  f.NoWait, err = ReadNoWait(reader)
  if err != nil:
    return errors.New("Error reading field NoWait")

}
func (f *BasicCancel) Write(writer io.Writer) {  err = WriteConsumerTag(writer)
  if err != nil:
    return errors.New("Error writing field ConsumerTag")

  err = WriteNoWait(writer)
  if err != nil:
    return errors.New("Error writing field NoWait")

}

// **********************************************************************
//                    Basic - CancelOk
// **********************************************************************

var MethodIdCancelOk uint16 = 31
type BasicCancelOk struct {
  ConsumerTag string
}
func (f *BasicCancelOk) Read(reader io.Reader) {  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil:
    return errors.New("Error reading field ConsumerTag")

}
func (f *BasicCancelOk) Write(writer io.Writer) {  err = WriteConsumerTag(writer)
  if err != nil:
    return errors.New("Error writing field ConsumerTag")

}

// **********************************************************************
//                    Basic - Publish
// **********************************************************************

var MethodIdPublish uint16 = 40
type BasicPublish struct {
  Reserved1 uint16
  Exchange string
  RoutingKey string
  Mandatory bool
  Immediate bool
}
func (f *BasicPublish) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field RoutingKey")

  f.Mandatory, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Mandatory")

  f.Immediate, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Immediate")

}
func (f *BasicPublish) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field RoutingKey")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Mandatory")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Immediate")

}

// **********************************************************************
//                    Basic - Return
// **********************************************************************

var MethodIdReturn uint16 = 50
type BasicReturn struct {
  ReplyCode uint16
  ReplyText string
  Exchange string
  RoutingKey string
}
func (f *BasicReturn) Read(reader io.Reader) {  f.ReplyCode, err = ReadReplyCode(reader)
  if err != nil:
    return errors.New("Error reading field ReplyCode")

  f.ReplyText, err = ReadReplyText(reader)
  if err != nil:
    return errors.New("Error reading field ReplyText")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field RoutingKey")

}
func (f *BasicReturn) Write(writer io.Writer) {  err = WriteReplyCode(writer)
  if err != nil:
    return errors.New("Error writing field ReplyCode")

  err = WriteReplyText(writer)
  if err != nil:
    return errors.New("Error writing field ReplyText")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field RoutingKey")

}

// **********************************************************************
//                    Basic - Deliver
// **********************************************************************

var MethodIdDeliver uint16 = 60
type BasicDeliver struct {
  ConsumerTag string
  DeliveryTag uint64
  Redelivered bool
  Exchange string
  RoutingKey string
}
func (f *BasicDeliver) Read(reader io.Reader) {  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil:
    return errors.New("Error reading field ConsumerTag")

  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil:
    return errors.New("Error reading field DeliveryTag")

  f.Redelivered, err = ReadRedelivered(reader)
  if err != nil:
    return errors.New("Error reading field Redelivered")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field RoutingKey")

}
func (f *BasicDeliver) Write(writer io.Writer) {  err = WriteConsumerTag(writer)
  if err != nil:
    return errors.New("Error writing field ConsumerTag")

  err = WriteDeliveryTag(writer)
  if err != nil:
    return errors.New("Error writing field DeliveryTag")

  err = WriteRedelivered(writer)
  if err != nil:
    return errors.New("Error writing field Redelivered")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field RoutingKey")

}

// **********************************************************************
//                    Basic - Get
// **********************************************************************

var MethodIdGet uint16 = 70
type BasicGet struct {
  Reserved1 uint16
  Queue string
  NoAck bool
}
func (f *BasicGet) Read(reader io.Reader) {  f.Reserved1, err = ReadShort(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

  f.Queue, err = ReadQueueName(reader)
  if err != nil:
    return errors.New("Error reading field Queue")

  f.NoAck, err = ReadNoAck(reader)
  if err != nil:
    return errors.New("Error reading field NoAck")

}
func (f *BasicGet) Write(writer io.Writer) {  err = WriteShort(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

  err = WriteQueueName(writer)
  if err != nil:
    return errors.New("Error writing field Queue")

  err = WriteNoAck(writer)
  if err != nil:
    return errors.New("Error writing field NoAck")

}

// **********************************************************************
//                    Basic - GetOk
// **********************************************************************

var MethodIdGetOk uint16 = 71
type BasicGetOk struct {
  DeliveryTag uint64
  Redelivered bool
  Exchange string
  RoutingKey string
  MessageCount uint32
}
func (f *BasicGetOk) Read(reader io.Reader) {  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil:
    return errors.New("Error reading field DeliveryTag")

  f.Redelivered, err = ReadRedelivered(reader)
  if err != nil:
    return errors.New("Error reading field Redelivered")

  f.Exchange, err = ReadExchangeName(reader)
  if err != nil:
    return errors.New("Error reading field Exchange")

  f.RoutingKey, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field RoutingKey")

  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil:
    return errors.New("Error reading field MessageCount")

}
func (f *BasicGetOk) Write(writer io.Writer) {  err = WriteDeliveryTag(writer)
  if err != nil:
    return errors.New("Error writing field DeliveryTag")

  err = WriteRedelivered(writer)
  if err != nil:
    return errors.New("Error writing field Redelivered")

  err = WriteExchangeName(writer)
  if err != nil:
    return errors.New("Error writing field Exchange")

  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field RoutingKey")

  err = WriteMessageCount(writer)
  if err != nil:
    return errors.New("Error writing field MessageCount")

}

// **********************************************************************
//                    Basic - GetEmpty
// **********************************************************************

var MethodIdGetEmpty uint16 = 72
type BasicGetEmpty struct {
  Reserved1 string
}
func (f *BasicGetEmpty) Read(reader io.Reader) {  f.Reserved1, err = ReadShortstr(reader)
  if err != nil:
    return errors.New("Error reading field Reserved1")

}
func (f *BasicGetEmpty) Write(writer io.Writer) {  err = WriteShortstr(writer)
  if err != nil:
    return errors.New("Error writing field Reserved1")

}

// **********************************************************************
//                    Basic - Ack
// **********************************************************************

var MethodIdAck uint16 = 80
type BasicAck struct {
  DeliveryTag uint64
  Multiple bool
}
func (f *BasicAck) Read(reader io.Reader) {  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil:
    return errors.New("Error reading field DeliveryTag")

  f.Multiple, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Multiple")

}
func (f *BasicAck) Write(writer io.Writer) {  err = WriteDeliveryTag(writer)
  if err != nil:
    return errors.New("Error writing field DeliveryTag")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Multiple")

}

// **********************************************************************
//                    Basic - Reject
// **********************************************************************

var MethodIdReject uint16 = 90
type BasicReject struct {
  DeliveryTag uint64
  Requeue bool
}
func (f *BasicReject) Read(reader io.Reader) {  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil:
    return errors.New("Error reading field DeliveryTag")

  f.Requeue, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Requeue")

}
func (f *BasicReject) Write(writer io.Writer) {  err = WriteDeliveryTag(writer)
  if err != nil:
    return errors.New("Error writing field DeliveryTag")

  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Requeue")

}

// **********************************************************************
//                    Basic - RecoverAsync
// **********************************************************************

var MethodIdRecoverAsync uint16 = 100
type BasicRecoverAsync struct {
  Requeue bool
}
func (f *BasicRecoverAsync) Read(reader io.Reader) {  f.Requeue, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Requeue")

}
func (f *BasicRecoverAsync) Write(writer io.Writer) {  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Requeue")

}

// **********************************************************************
//                    Basic - Recover
// **********************************************************************

var MethodIdRecover uint16 = 110
type BasicRecover struct {
  Requeue bool
}
func (f *BasicRecover) Read(reader io.Reader) {  f.Requeue, err = ReadBit(reader)
  if err != nil:
    return errors.New("Error reading field Requeue")

}
func (f *BasicRecover) Write(writer io.Writer) {  err = WriteBit(writer)
  if err != nil:
    return errors.New("Error writing field Requeue")

}

// **********************************************************************
//                    Basic - RecoverOk
// **********************************************************************

var MethodIdRecoverOk uint16 = 111
type BasicRecoverOk struct {
}
func (f *BasicRecoverOk) Read(reader io.Reader) {}
func (f *BasicRecoverOk) Write(writer io.Writer) {}

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

var MethodIdSelect uint16 = 10
type TxSelect struct {
}
func (f *TxSelect) Read(reader io.Reader) {}
func (f *TxSelect) Write(writer io.Writer) {}

// **********************************************************************
//                    Tx - SelectOk
// **********************************************************************

var MethodIdSelectOk uint16 = 11
type TxSelectOk struct {
}
func (f *TxSelectOk) Read(reader io.Reader) {}
func (f *TxSelectOk) Write(writer io.Writer) {}

// **********************************************************************
//                    Tx - Commit
// **********************************************************************

var MethodIdCommit uint16 = 20
type TxCommit struct {
}
func (f *TxCommit) Read(reader io.Reader) {}
func (f *TxCommit) Write(writer io.Writer) {}

// **********************************************************************
//                    Tx - CommitOk
// **********************************************************************

var MethodIdCommitOk uint16 = 21
type TxCommitOk struct {
}
func (f *TxCommitOk) Read(reader io.Reader) {}
func (f *TxCommitOk) Write(writer io.Writer) {}

// **********************************************************************
//                    Tx - Rollback
// **********************************************************************

var MethodIdRollback uint16 = 30
type TxRollback struct {
}
func (f *TxRollback) Read(reader io.Reader) {}
func (f *TxRollback) Write(writer io.Writer) {}

// **********************************************************************
//                    Tx - RollbackOk
// **********************************************************************

var MethodIdRollbackOk uint16 = 31
type TxRollbackOk struct {
}
func (f *TxRollbackOk) Read(reader io.Reader) {}
func (f *TxRollbackOk) Write(writer io.Writer) {}
