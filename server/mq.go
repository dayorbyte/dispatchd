
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
  if err := startConnection(conn); err != nil {
    fmt.Println(err)
    panic("Error opening connection")
  }
  fmt.Println("Connection open!")
  fmt.Println("Reading frames and throwing them away")
  for {
    frame, err := amqp.ReadFrame(conn)
    _ = frame
    if err != nil {
      panic("bad frame!")
    }
  }
}

func openOk(conn net.Conn) error {
  fmt.Println("=====> Sending 'openOk'")
  send(conn, &amqp.ConnectionOpenOk{""})
  return openOk(conn)
}

func open(conn net.Conn, startOk *amqp.ConnectionTuneOk) error {
  fmt.Println("=====> Receiving 'open'")
  method, err := receive(conn)
  if err != nil {
    return err
  }
  _, rightType := method.(*amqp.ConnectionOpen)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected open")
  }
  return nil
}

func tuneOk(conn net.Conn) error {
  fmt.Println("=====> Receiving 'tuneOk'")
  method, err := receive(conn)
  if err != nil {
    return err
  }

  tok, rightType := method.(*amqp.ConnectionTuneOk)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected tunOk")
  }
  return open(conn, tok)
}

func tune(conn net.Conn, startOk *amqp.ConnectionStartOk) error {
  fmt.Println("=====> Sending 'tune'")
  send(conn, &amqp.ConnectionTune{256, 8192, 10})
  return tuneOk(conn)
}

func startOk(conn net.Conn) error {
  fmt.Println("=====> Receiving 'startOk'")
  method, err := receive(conn)
  if err != nil {
    return err
  }

  sok, rightType := method.(*amqp.ConnectionStartOk)
  if !rightType {
    return errors.New("Got the wrong frame type. Expected startOk")
  }
  return tune(conn, sok)
}

func startConnection(conn net.Conn) error {
  fmt.Println("=====> Sending 'start'")
  send(conn, &amqp.ConnectionStart{0, 9, amqp.Table{}, []byte("PLAIN"), []byte("en_US")})
  return startOk(conn)
}

func send(conn net.Conn, method amqp.MethodFrame) error {
  var buf = bytes.NewBuffer([]byte{})
  method.Write(buf)
  var frame = &amqp.FrameWrapper{uint8(amqp.FrameMethod), 0, buf.Bytes()}
  amqp.WriteFrame(conn, frame)
  return nil
}

func receive(conn net.Conn) (amqp.MethodFrame, error) {
  for {
    frame, err := amqp.ReadFrame(conn)
    if err != nil {
      return nil, err
    }
    fmt.Println("Frame info", frame.FrameType, frame.Channel)
    var methodReader = bytes.NewReader(frame.Payload)
    return amqp.ReadMethod(methodReader)
  }

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
