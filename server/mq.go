
package main

import (
  "fmt"
  "net"
  "os"
  "bytes"
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
  fmt.Println("Got bytes: %d", length)

}

func start(conn net.Conn) {
  var buf = bytes.NewBuffer([]byte{})
  // Method headers
  WriterMethodHeader(buf, ClassConnection, MethodStart)

  // Protocol version
  WriteVersion(buf)

  // Server properties
  WriteServerProps(buf)

  // Mechanisms
  WriteLongStr(buf, "PLAIN")

  // Locals
  WriteLongStr(buf, "en_US")

  // Send to client
  WriteFrameHeader(conn, FrameMethod, 0, uint32(buf.Len()))
  conn.Write(buf.Bytes())
  WriteFrameEnd(conn)
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
