package amqp

import (
  //"encoding/binary"
  "io"
  "errors"
  //"bytes"
  "strconv"
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


func (f* ConnectionStart) MethodIdentifier() (uint16, uint16) {
  return 10, 10
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
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
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


func (f* ConnectionStartOk) MethodIdentifier() (uint16, uint16) {
  return 10, 11
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
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 11); err != nil {
    return err
  }
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


func (f* ConnectionSecure) MethodIdentifier() (uint16, uint16) {
  return 10, 20
}

func (f *ConnectionSecure) Read(reader io.Reader) (err error) {
  f.Challenge, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Challenge")
  }

  return
}
func (f *ConnectionSecure) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
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


func (f* ConnectionSecureOk) MethodIdentifier() (uint16, uint16) {
  return 10, 21
}

func (f *ConnectionSecureOk) Read(reader io.Reader) (err error) {
  f.Response, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Response")
  }

  return
}
func (f *ConnectionSecureOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 21); err != nil {
    return err
  }
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


func (f* ConnectionTune) MethodIdentifier() (uint16, uint16) {
  return 10, 30
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
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 30); err != nil {
    return err
  }
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


func (f* ConnectionTuneOk) MethodIdentifier() (uint16, uint16) {
  return 10, 31
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
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 31); err != nil {
    return err
  }
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


func (f* ConnectionOpen) MethodIdentifier() (uint16, uint16) {
  return 10, 40
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Reserved2 = (bits & (1 << 0) > 0)

  return
}
func (f *ConnectionOpen) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
  err = WritePath(writer, f.VirtualHost)
  if err != nil {
    return errors.New("Error writing field VirtualHost")
  }

  err = WriteShortstr(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  var bits byte
  if f.Reserved2 {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Connection - OpenOk
// **********************************************************************

var MethodIdConnectionOpenOk uint16 = 41
type ConnectionOpenOk struct {
  Reserved1 string
}


func (f* ConnectionOpenOk) MethodIdentifier() (uint16, uint16) {
  return 10, 41
}

func (f *ConnectionOpenOk) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *ConnectionOpenOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 41); err != nil {
    return err
  }
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


func (f* ConnectionClose) MethodIdentifier() (uint16, uint16) {
  return 10, 50
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
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
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


func (f* ConnectionCloseOk) MethodIdentifier() (uint16, uint16) {
  return 10, 51
}

func (f *ConnectionCloseOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ConnectionCloseOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  if err = WriteShort(writer, 51); err != nil {
    return err
  }
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


func (f* ChannelOpen) MethodIdentifier() (uint16, uint16) {
  return 20, 10
}

func (f *ChannelOpen) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *ChannelOpen) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
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


func (f* ChannelOpenOk) MethodIdentifier() (uint16, uint16) {
  return 20, 11
}

func (f *ChannelOpenOk) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadLongstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *ChannelOpenOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  if err = WriteShort(writer, 11); err != nil {
    return err
  }
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


func (f* ChannelFlow) MethodIdentifier() (uint16, uint16) {
  return 20, 20
}

func (f *ChannelFlow) Read(reader io.Reader) (err error) {

  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Active = (bits & (1 << 0) > 0)

  return
}
func (f *ChannelFlow) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  var bits byte
  if f.Active {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Channel - FlowOk
// **********************************************************************

var MethodIdChannelFlowOk uint16 = 21
type ChannelFlowOk struct {
  Active bool
}


func (f* ChannelFlowOk) MethodIdentifier() (uint16, uint16) {
  return 20, 21
}

func (f *ChannelFlowOk) Read(reader io.Reader) (err error) {

  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Active = (bits & (1 << 0) > 0)

  return
}
func (f *ChannelFlowOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  if err = WriteShort(writer, 21); err != nil {
    return err
  }
  var bits byte
  if f.Active {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* ChannelClose) MethodIdentifier() (uint16, uint16) {
  return 20, 40
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
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
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


func (f* ChannelCloseOk) MethodIdentifier() (uint16, uint16) {
  return 20, 41
}

func (f *ChannelCloseOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ChannelCloseOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  if err = WriteShort(writer, 41); err != nil {
    return err
  }
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


func (f* ExchangeDeclare) MethodIdentifier() (uint16, uint16) {
  return 40, 10
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Passive = (bits & (1 << 0) > 0)

  f.Durable = (bits & (1 << 1) > 0)

  f.Reserved2 = (bits & (1 << 2) > 0)

  f.Reserved3 = (bits & (1 << 3) > 0)

  f.NoWait = (bits & (1 << 4) > 0)

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *ExchangeDeclare) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
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

  var bits byte
  if f.Passive {
    bits |= 1 << 0
  }
  if f.Durable {
    bits |= 1 << 1
  }
  if f.Reserved2 {
    bits |= 1 << 2
  }
  if f.Reserved3 {
    bits |= 1 << 3
  }
  if f.NoWait {
    bits |= 1 << 4
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* ExchangeDeclareOk) MethodIdentifier() (uint16, uint16) {
  return 40, 11
}

func (f *ExchangeDeclareOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ExchangeDeclareOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
  if err = WriteShort(writer, 11); err != nil {
    return err
  }
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


func (f* ExchangeDelete) MethodIdentifier() (uint16, uint16) {
  return 40, 20
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.IfUnused = (bits & (1 << 0) > 0)

  f.NoWait = (bits & (1 << 1) > 0)

  return
}
func (f *ExchangeDelete) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  err = WriteExchangeName(writer, f.Exchange)
  if err != nil {
    return errors.New("Error writing field Exchange")
  }

  var bits byte
  if f.IfUnused {
    bits |= 1 << 0
  }
  if f.NoWait {
    bits |= 1 << 1
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Exchange - DeleteOk
// **********************************************************************

var MethodIdExchangeDeleteOk uint16 = 21
type ExchangeDeleteOk struct {
}


func (f* ExchangeDeleteOk) MethodIdentifier() (uint16, uint16) {
  return 40, 21
}

func (f *ExchangeDeleteOk) Read(reader io.Reader) (err error) {
  return
}
func (f *ExchangeDeleteOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
  if err = WriteShort(writer, 21); err != nil {
    return err
  }
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


func (f* QueueDeclare) MethodIdentifier() (uint16, uint16) {
  return 50, 10
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Passive = (bits & (1 << 0) > 0)

  f.Durable = (bits & (1 << 1) > 0)

  f.Exclusive = (bits & (1 << 2) > 0)

  f.AutoDelete = (bits & (1 << 3) > 0)

  f.NoWait = (bits & (1 << 4) > 0)

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *QueueDeclare) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

  var bits byte
  if f.Passive {
    bits |= 1 << 0
  }
  if f.Durable {
    bits |= 1 << 1
  }
  if f.Exclusive {
    bits |= 1 << 2
  }
  if f.AutoDelete {
    bits |= 1 << 3
  }
  if f.NoWait {
    bits |= 1 << 4
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* QueueDeclareOk) MethodIdentifier() (uint16, uint16) {
  return 50, 11
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
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 11); err != nil {
    return err
  }
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


func (f* QueueBind) MethodIdentifier() (uint16, uint16) {
  return 50, 20
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.NoWait = (bits & (1 << 0) > 0)

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *QueueBind) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
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

  var bits byte
  if f.NoWait {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* QueueBindOk) MethodIdentifier() (uint16, uint16) {
  return 50, 21
}

func (f *QueueBindOk) Read(reader io.Reader) (err error) {
  return
}
func (f *QueueBindOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 21); err != nil {
    return err
  }
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


func (f* QueueUnbind) MethodIdentifier() (uint16, uint16) {
  return 50, 50
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
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
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


func (f* QueueUnbindOk) MethodIdentifier() (uint16, uint16) {
  return 50, 51
}

func (f *QueueUnbindOk) Read(reader io.Reader) (err error) {
  return
}
func (f *QueueUnbindOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 51); err != nil {
    return err
  }
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


func (f* QueuePurge) MethodIdentifier() (uint16, uint16) {
  return 50, 30
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.NoWait = (bits & (1 << 0) > 0)

  return
}
func (f *QueuePurge) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 30); err != nil {
    return err
  }
  err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

  var bits byte
  if f.NoWait {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Queue - PurgeOk
// **********************************************************************

var MethodIdQueuePurgeOk uint16 = 31
type QueuePurgeOk struct {
  MessageCount uint32
}


func (f* QueuePurgeOk) MethodIdentifier() (uint16, uint16) {
  return 50, 31
}

func (f *QueuePurgeOk) Read(reader io.Reader) (err error) {
  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil {
    return errors.New("Error reading field MessageCount")
  }

  return
}
func (f *QueuePurgeOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 31); err != nil {
    return err
  }
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


func (f* QueueDelete) MethodIdentifier() (uint16, uint16) {
  return 50, 40
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.IfUnused = (bits & (1 << 0) > 0)

  f.IfEmpty = (bits & (1 << 1) > 0)

  f.NoWait = (bits & (1 << 2) > 0)

  return
}
func (f *QueueDelete) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
  err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

  var bits byte
  if f.IfUnused {
    bits |= 1 << 0
  }
  if f.IfEmpty {
    bits |= 1 << 1
  }
  if f.NoWait {
    bits |= 1 << 2
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Queue - DeleteOk
// **********************************************************************

var MethodIdQueueDeleteOk uint16 = 41
type QueueDeleteOk struct {
  MessageCount uint32
}


func (f* QueueDeleteOk) MethodIdentifier() (uint16, uint16) {
  return 50, 41
}

func (f *QueueDeleteOk) Read(reader io.Reader) (err error) {
  f.MessageCount, err = ReadMessageCount(reader)
  if err != nil {
    return errors.New("Error reading field MessageCount")
  }

  return
}
func (f *QueueDeleteOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
  if err = WriteShort(writer, 41); err != nil {
    return err
  }
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


func (f* BasicQos) MethodIdentifier() (uint16, uint16) {
  return 60, 10
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Global = (bits & (1 << 0) > 0)

  return
}
func (f *BasicQos) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  err = WriteLong(writer, f.PrefetchSize)
  if err != nil {
    return errors.New("Error writing field PrefetchSize")
  }

  err = WriteShort(writer, f.PrefetchCount)
  if err != nil {
    return errors.New("Error writing field PrefetchCount")
  }

  var bits byte
  if f.Global {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Basic - QosOk
// **********************************************************************

var MethodIdBasicQosOk uint16 = 11
type BasicQosOk struct {
}


func (f* BasicQosOk) MethodIdentifier() (uint16, uint16) {
  return 60, 11
}

func (f *BasicQosOk) Read(reader io.Reader) (err error) {
  return
}
func (f *BasicQosOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 11); err != nil {
    return err
  }
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


func (f* BasicConsume) MethodIdentifier() (uint16, uint16) {
  return 60, 20
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.NoLocal = (bits & (1 << 0) > 0)

  f.NoAck = (bits & (1 << 1) > 0)

  f.Exclusive = (bits & (1 << 2) > 0)

  f.NoWait = (bits & (1 << 3) > 0)

  f.Arguments, err = ReadTable(reader)
  if err != nil {
    return errors.New("Error reading field Arguments")
  }

  return
}
func (f *BasicConsume) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
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

  var bits byte
  if f.NoLocal {
    bits |= 1 << 0
  }
  if f.NoAck {
    bits |= 1 << 1
  }
  if f.Exclusive {
    bits |= 1 << 2
  }
  if f.NoWait {
    bits |= 1 << 3
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* BasicConsumeOk) MethodIdentifier() (uint16, uint16) {
  return 60, 21
}

func (f *BasicConsumeOk) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  return
}
func (f *BasicConsumeOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 21); err != nil {
    return err
  }
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


func (f* BasicCancel) MethodIdentifier() (uint16, uint16) {
  return 60, 30
}

func (f *BasicCancel) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.NoWait = (bits & (1 << 0) > 0)

  return
}
func (f *BasicCancel) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 30); err != nil {
    return err
  }
  err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

  var bits byte
  if f.NoWait {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Basic - CancelOk
// **********************************************************************

var MethodIdBasicCancelOk uint16 = 31
type BasicCancelOk struct {
  ConsumerTag string
}


func (f* BasicCancelOk) MethodIdentifier() (uint16, uint16) {
  return 60, 31
}

func (f *BasicCancelOk) Read(reader io.Reader) (err error) {
  f.ConsumerTag, err = ReadConsumerTag(reader)
  if err != nil {
    return errors.New("Error reading field ConsumerTag")
  }

  return
}
func (f *BasicCancelOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 31); err != nil {
    return err
  }
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


func (f* BasicPublish) MethodIdentifier() (uint16, uint16) {
  return 60, 40
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Mandatory = (bits & (1 << 0) > 0)

  f.Immediate = (bits & (1 << 1) > 0)

  return
}
func (f *BasicPublish) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 40); err != nil {
    return err
  }
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

  var bits byte
  if f.Mandatory {
    bits |= 1 << 0
  }
  if f.Immediate {
    bits |= 1 << 1
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* BasicReturn) MethodIdentifier() (uint16, uint16) {
  return 60, 50
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
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 50); err != nil {
    return err
  }
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


func (f* BasicDeliver) MethodIdentifier() (uint16, uint16) {
  return 60, 60
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Redelivered = (bits & (1 << 0) > 0)

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
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  err = WriteConsumerTag(writer, f.ConsumerTag)
  if err != nil {
    return errors.New("Error writing field ConsumerTag")
  }

  err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

  var bits byte
  if f.Redelivered {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* BasicGet) MethodIdentifier() (uint16, uint16) {
  return 60, 70
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


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.NoAck = (bits & (1 << 0) > 0)

  return
}
func (f *BasicGet) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 70); err != nil {
    return err
  }
  err = WriteShort(writer, f.Reserved1)
  if err != nil {
    return errors.New("Error writing field Reserved1")
  }

  err = WriteQueueName(writer, f.Queue)
  if err != nil {
    return errors.New("Error writing field Queue")
  }

  var bits byte
  if f.NoAck {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* BasicGetOk) MethodIdentifier() (uint16, uint16) {
  return 60, 71
}

func (f *BasicGetOk) Read(reader io.Reader) (err error) {
  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Redelivered = (bits & (1 << 0) > 0)

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
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 71); err != nil {
    return err
  }
  err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

  var bits byte
  if f.Redelivered {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* BasicGetEmpty) MethodIdentifier() (uint16, uint16) {
  return 60, 72
}

func (f *BasicGetEmpty) Read(reader io.Reader) (err error) {
  f.Reserved1, err = ReadShortstr(reader)
  if err != nil {
    return errors.New("Error reading field Reserved1")
  }

  return
}
func (f *BasicGetEmpty) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 72); err != nil {
    return err
  }
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


func (f* BasicAck) MethodIdentifier() (uint16, uint16) {
  return 60, 80
}

func (f *BasicAck) Read(reader io.Reader) (err error) {
  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Multiple = (bits & (1 << 0) > 0)

  return
}
func (f *BasicAck) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 80); err != nil {
    return err
  }
  err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

  var bits byte
  if f.Multiple {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

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


func (f* BasicReject) MethodIdentifier() (uint16, uint16) {
  return 60, 90
}

func (f *BasicReject) Read(reader io.Reader) (err error) {
  f.DeliveryTag, err = ReadDeliveryTag(reader)
  if err != nil {
    return errors.New("Error reading field DeliveryTag")
  }


  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Requeue = (bits & (1 << 0) > 0)

  return
}
func (f *BasicReject) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  err = WriteDeliveryTag(writer, f.DeliveryTag)
  if err != nil {
    return errors.New("Error writing field DeliveryTag")
  }

  var bits byte
  if f.Requeue {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Basic - RecoverAsync
// **********************************************************************

var MethodIdBasicRecoverAsync uint16 = 100
type BasicRecoverAsync struct {
  Requeue bool
}


func (f* BasicRecoverAsync) MethodIdentifier() (uint16, uint16) {
  return 60, 100
}

func (f *BasicRecoverAsync) Read(reader io.Reader) (err error) {

  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Requeue = (bits & (1 << 0) > 0)

  return
}
func (f *BasicRecoverAsync) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 100); err != nil {
    return err
  }
  var bits byte
  if f.Requeue {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Basic - Recover
// **********************************************************************

var MethodIdBasicRecover uint16 = 110
type BasicRecover struct {
  Requeue bool
}


func (f* BasicRecover) MethodIdentifier() (uint16, uint16) {
  return 60, 110
}

func (f *BasicRecover) Read(reader io.Reader) (err error) {

  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}


  f.Requeue = (bits & (1 << 0) > 0)

  return
}
func (f *BasicRecover) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 110); err != nil {
    return err
  }
  var bits byte
  if f.Requeue {
    bits |= 1 << 0
  }
  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}

  return
}

// **********************************************************************
//                    Basic - RecoverOk
// **********************************************************************

var MethodIdBasicRecoverOk uint16 = 111
type BasicRecoverOk struct {
}


func (f* BasicRecoverOk) MethodIdentifier() (uint16, uint16) {
  return 60, 111
}

func (f *BasicRecoverOk) Read(reader io.Reader) (err error) {
  return
}
func (f *BasicRecoverOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 60); err != nil {
    return err
  }
  if err = WriteShort(writer, 111); err != nil {
    return err
  }
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


func (f* TxSelect) MethodIdentifier() (uint16, uint16) {
  return 90, 10
}

func (f *TxSelect) Read(reader io.Reader) (err error) {
  return
}
func (f *TxSelect) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  if err = WriteShort(writer, 10); err != nil {
    return err
  }
  return
}

// **********************************************************************
//                    Tx - SelectOk
// **********************************************************************

var MethodIdTxSelectOk uint16 = 11
type TxSelectOk struct {
}


func (f* TxSelectOk) MethodIdentifier() (uint16, uint16) {
  return 90, 11
}

func (f *TxSelectOk) Read(reader io.Reader) (err error) {
  return
}
func (f *TxSelectOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  if err = WriteShort(writer, 11); err != nil {
    return err
  }
  return
}

// **********************************************************************
//                    Tx - Commit
// **********************************************************************

var MethodIdTxCommit uint16 = 20
type TxCommit struct {
}


func (f* TxCommit) MethodIdentifier() (uint16, uint16) {
  return 90, 20
}

func (f *TxCommit) Read(reader io.Reader) (err error) {
  return
}
func (f *TxCommit) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  if err = WriteShort(writer, 20); err != nil {
    return err
  }
  return
}

// **********************************************************************
//                    Tx - CommitOk
// **********************************************************************

var MethodIdTxCommitOk uint16 = 21
type TxCommitOk struct {
}


func (f* TxCommitOk) MethodIdentifier() (uint16, uint16) {
  return 90, 21
}

func (f *TxCommitOk) Read(reader io.Reader) (err error) {
  return
}
func (f *TxCommitOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  if err = WriteShort(writer, 21); err != nil {
    return err
  }
  return
}

// **********************************************************************
//                    Tx - Rollback
// **********************************************************************

var MethodIdTxRollback uint16 = 30
type TxRollback struct {
}


func (f* TxRollback) MethodIdentifier() (uint16, uint16) {
  return 90, 30
}

func (f *TxRollback) Read(reader io.Reader) (err error) {
  return
}
func (f *TxRollback) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  if err = WriteShort(writer, 30); err != nil {
    return err
  }
  return
}

// **********************************************************************
//                    Tx - RollbackOk
// **********************************************************************

var MethodIdTxRollbackOk uint16 = 31
type TxRollbackOk struct {
}


func (f* TxRollbackOk) MethodIdentifier() (uint16, uint16) {
  return 90, 31
}

func (f *TxRollbackOk) Read(reader io.Reader) (err error) {
  return
}
func (f *TxRollbackOk) Write(writer io.Writer) (err error) {
  if err = WriteShort(writer, 90); err != nil {
    return err
  }
  if err = WriteShort(writer, 31); err != nil {
    return err
  }
  return
}
func ReadMethod(reader io.Reader) (MethodFrame, error) {
  classIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  methodIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  switch {
    // Connection
    case classIndex == 10:
      switch {
      case methodIndex == 10: // ConnectionStart
        var method = &ConnectionStart{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 11: // ConnectionStartOk
        var method = &ConnectionStartOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 20: // ConnectionSecure
        var method = &ConnectionSecure{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 21: // ConnectionSecureOk
        var method = &ConnectionSecureOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 30: // ConnectionTune
        var method = &ConnectionTune{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 31: // ConnectionTuneOk
        var method = &ConnectionTuneOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 40: // ConnectionOpen
        var method = &ConnectionOpen{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 41: // ConnectionOpenOk
        var method = &ConnectionOpenOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 50: // ConnectionClose
        var method = &ConnectionClose{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 51: // ConnectionCloseOk
        var method = &ConnectionCloseOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
    }
    // Channel
    case classIndex == 20:
      switch {
      case methodIndex == 10: // ChannelOpen
        var method = &ChannelOpen{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 11: // ChannelOpenOk
        var method = &ChannelOpenOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 20: // ChannelFlow
        var method = &ChannelFlow{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 21: // ChannelFlowOk
        var method = &ChannelFlowOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 40: // ChannelClose
        var method = &ChannelClose{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 41: // ChannelCloseOk
        var method = &ChannelCloseOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
    }
    // Exchange
    case classIndex == 40:
      switch {
      case methodIndex == 10: // ExchangeDeclare
        var method = &ExchangeDeclare{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 11: // ExchangeDeclareOk
        var method = &ExchangeDeclareOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 20: // ExchangeDelete
        var method = &ExchangeDelete{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 21: // ExchangeDeleteOk
        var method = &ExchangeDeleteOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
    }
    // Queue
    case classIndex == 50:
      switch {
      case methodIndex == 10: // QueueDeclare
        var method = &QueueDeclare{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 11: // QueueDeclareOk
        var method = &QueueDeclareOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 20: // QueueBind
        var method = &QueueBind{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 21: // QueueBindOk
        var method = &QueueBindOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 30: // QueuePurge
        var method = &QueuePurge{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 31: // QueuePurgeOk
        var method = &QueuePurgeOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 40: // QueueDelete
        var method = &QueueDelete{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 41: // QueueDeleteOk
        var method = &QueueDeleteOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 50: // QueueUnbind
        var method = &QueueUnbind{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 51: // QueueUnbindOk
        var method = &QueueUnbindOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
    }
    // Basic
    case classIndex == 60:
      switch {
      case methodIndex == 10: // BasicQos
        var method = &BasicQos{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 100: // BasicRecoverAsync
        var method = &BasicRecoverAsync{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 11: // BasicQosOk
        var method = &BasicQosOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 110: // BasicRecover
        var method = &BasicRecover{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 111: // BasicRecoverOk
        var method = &BasicRecoverOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 20: // BasicConsume
        var method = &BasicConsume{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 21: // BasicConsumeOk
        var method = &BasicConsumeOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 30: // BasicCancel
        var method = &BasicCancel{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 31: // BasicCancelOk
        var method = &BasicCancelOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 40: // BasicPublish
        var method = &BasicPublish{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 50: // BasicReturn
        var method = &BasicReturn{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 60: // BasicDeliver
        var method = &BasicDeliver{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 70: // BasicGet
        var method = &BasicGet{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 71: // BasicGetOk
        var method = &BasicGetOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 72: // BasicGetEmpty
        var method = &BasicGetEmpty{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 80: // BasicAck
        var method = &BasicAck{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 90: // BasicReject
        var method = &BasicReject{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
    }
    // Tx
    case classIndex == 90:
      switch {
      case methodIndex == 10: // TxSelect
        var method = &TxSelect{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 11: // TxSelectOk
        var method = &TxSelectOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 20: // TxCommit
        var method = &TxCommit{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 21: // TxCommitOk
        var method = &TxCommitOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 30: // TxRollback
        var method = &TxRollback{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
      case methodIndex == 31: // TxRollbackOk
        var method = &TxRollbackOk{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil
    }
  }
  return nil, errors.New(
    "Bad method or class Id! classId:" +
    strconv.FormatUint(uint64(classIndex), 10) +
    " methodIndex: " +
    strconv.FormatUint(uint64(methodIndex), 10))
}