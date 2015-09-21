
package main

import (
  "fmt"
  "net"
  "os"
  "bytes"
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
  fmt.Println("Sending StartOk")
  start(conn)
  fmt.Println("Waiting for bytes")
  buf2 := make([]byte, 10000)
  length, err := conn.Read(buf2)
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }
  fmt.Println("Got bytes:", length)

}

func start(conn net.Conn) {
  var buf = bytes.NewBuffer([]byte{})
  // Method headers
  amqp.WriteShort(buf, amqp.ClassIdConnection)
  amqp.WriteShort(buf, amqp.MethodIdConnectionStart)

  // Protocol version
  amqp.WriteVersion(buf)

  // Server properties
  amqp.WriteTable(buf, amqp.Table{})

  // Mechanisms
  amqp.WriteLongstr(buf, []byte("PLAIN"))

  // Locals
  amqp.WriteLongstr(buf, []byte("en_US"))

  // Send to client
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
