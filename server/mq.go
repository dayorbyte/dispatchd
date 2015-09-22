
package main

import (
  "fmt"
  "net"
  "errors"
  "os"
  "bytes"
  "github.com/jeffjenkins/mq/amqp"
)

func handleConnection(conn net.Conn) {
  fmt.Println("=====> Incoming connection!")
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
  if startConnection(conn) != nil {
    panic("Error opening connection")
  }
  fmt.Println("Connection open!")
}

func tuneOk(conn net.Conn) error {
  fmt.Println("=====> Receiving 'tuneOk'")
  panic("Not implemented!")
}

func tune(conn net.Conn, startOk *amqp.ConnectionStartOk) error {
  fmt.Println("=====> Sending 'tune'")
  panic("Implement tune!")

  // Method
  var buf = bytes.NewBuffer([]byte{})
  var tune = amqp.ConnectionTune{

  }
  tune.Write(buf)

  // Frame
  var frame = &amqp.FrameWrapper{uint8(amqp.FrameMethod), 0, buf.Bytes()}
  amqp.WriteFrame(conn, frame)

  // Receive tuneOk
  return tuneOk(conn)
}

func startOk(conn net.Conn) error {
  fmt.Println("=====> Receiving 'startOk'")
  frame, err := amqp.ReadFrame(conn)
  if err != nil {
    return err
  }
  var methodReader = bytes.NewReader(frame.Payload)
  method, err := amqp.ReadMethod(methodReader)
  if err != nil {
    return err
  }
  sok, rightType := method.(*amqp.ConnectionStartOk)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected startOk")
  }
  // for key := range ok.ClientProperties {
  //   fmt.Println("ClientProperties", key)
  // }
  // fmt.Println("Mechanism"             , ok.Mechanism)
  // fmt.Println("Response"              , ok.Response, string(ok.Response))
  // fmt.Println("Locale"                , ok.Locale)
  return tune(conn, sok)
}

func startConnection(conn net.Conn) error {
  fmt.Println("=====> Sending 'start'")
  // Method
  var buf = bytes.NewBuffer([]byte{})
  var start = amqp.ConnectionStart{0, 9, amqp.Table{}, []byte("PLAIN"), []byte("en_US")}
  start.Write(buf)

  // Frame
  var frame = &amqp.FrameWrapper{uint8(amqp.FrameMethod), 0, buf.Bytes()}
  amqp.WriteFrame(conn, frame)

  // Receive startOk
  return startOk(conn)
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
