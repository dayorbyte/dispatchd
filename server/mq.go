
package main

import (
  "fmt"
  "net"
  "os"
)

func handleConnection(conn net.Conn) {
  c := NewAMQPConnection(conn)
  c.openConnection()
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
