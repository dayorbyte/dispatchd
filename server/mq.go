
package main

import (
  "fmt"
  "net"
  "os"
  "bytes"
  "encoding/binary"
  "github.com/jeffjenkins/mq/amqp"
)

func handleConnection(conn net.Conn) {
  fmt.Println("Got connection!")
  // Negotiate Protocol
  buf := make([]byte, 8)
  _, err := conn.Read(buf)
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }
  var supported = []byte { 'A', 'M', 'Q', 'P', 0, 0, 9, 1 }
  if bytes.Compare(buf, supported) != 0 {
    conn.Write(supported)
    conn.Close()
    return
  }
  fmt.Println("Connection open!")
  // Connection is open!
  fmt.Println("Sending Start")
  start(conn)
  var ok, sokerr = startOk(conn)
  if sokerr != nil {
    panic("Error in startOK!")
  }
  for key := range ok.ClientProperties {
    fmt.Println("ClientProperties", key)
  }

  fmt.Println("Mechanism"             , ok.Mechanism)
  fmt.Println("Response"              , ok.Response, string(ok.Response))
  fmt.Println("Locale"                , ok.Locale)
}

func startOk(conn net.Conn) (amqp.ConnectionStartOk, error) {
  var ftype, _ = amqp.ReadOctet(conn) // frame type
  fmt.Println("ftype: ", ftype)

  var channel, _ = amqp.ReadShort(conn) // channel
  fmt.Println("channel: ", channel)

  var length, _ = amqp.ReadLong(conn)  // paload length
  fmt.Println("length: ", length)

  var payload = make([]byte, length)
  fmt.Println("Reading payload")
  binary.Read(conn, binary.BigEndian, &payload)

  var methodReader = bytes.NewReader(payload)
  fmt.Println("Payload length:", len(payload))

  amqp.ReadShort(methodReader) // ClassId
  amqp.ReadShort(methodReader) // MethodId

  var method = amqp.ConnectionStartOk{}
  method.Read(methodReader)
  return method, nil
}

func start(conn net.Conn) {
  var buf = bytes.NewBuffer([]byte{})
  // Method headers

  amqp.WriteShort(buf, amqp.ClassIdConnection)
  amqp.WriteShort(buf, amqp.MethodIdConnectionStart)
  var start = amqp.ConnectionStart{0, 9, amqp.Table{}, []byte("PLAIN"), []byte("en_US")}
  start.Write(buf)

  // Frame header
  amqp.WriteOctet(conn, uint8(amqp.FrameMethod))
  amqp.WriteShort(conn, 0)
  amqp.WriteLong(conn, uint32(buf.Len()))
  conn.Write(buf.Bytes())
  amqp.WriteFrameEnd(conn)
}

func main() {
  fmt.Printf("Listening on port 1111\n")
  ln, err := net.Listen("tcp", ":1111")
  if err != nil {
    fmt.Printf("Error!")
    os.Exit(1)
  }
  for {
    conn, err := ln.Accept()
    if err != nil {
      fmt.Printf("Error accepting connection!\n")
      os.Exit(1)
    }
    go handleConnection(conn)
  }
}
