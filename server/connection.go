
package main

import (
  "fmt"
  "net"
  "errors"
  // "os"
  "bytes"
  "github.com/jeffjenkins/mq/amqp"
)

type AMQPConnection struct {
  network net.Conn
  open bool
  done bool
  // this is thread-safe because only channel 0 manages these fields
  nextChannel int
  channels map[uint16]Channel

}

func NewAMQPConnection(network net.Conn) *AMQPConnection {
  return &AMQPConnection{
    network,
    false,
    false,
    0,
    make(map[uint16]Channel),
  }
}

func (conn *AMQPConnection) openConnection() error {
    fmt.Println("=====> Incoming connection!")
  // Negotiate Protocol
  buf := make([]byte, 8)
  _, err := conn.network.Read(buf)
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }
  var supported = []byte { 'A', 'M', 'Q', 'P', 0, 0, 9, 1 }
  if bytes.Compare(buf, supported) != 0 {
    conn.network.Write(supported)
    conn.network.Close()
    return errors.New("Bad version from client")
  }
  if err := conn.startConnection(); err != nil {
    fmt.Println(err)
    panic("Error opening connection")
  }
  return conn.handleFrames()
}

func (conn *AMQPConnection) handleFrames() error {
  fmt.Println("Connection open!")
  fmt.Println("Reading frames and throwing them away")
  for {
    // If the connection is done, we stop handling frames
    if conn.done {
      break
    }
    frame, err := amqp.ReadFrame(conn.network)
    if err != nil {
      return err
    }
    // If we haven't finished the handshake, ignore frames on channels other
    // than 0
    if !conn.open && frame.Channel != 0 {
      continue
    }
    var channel, ok = conn.channels[frame.Channel]
    if !ok {
      continue
    }
    // Dispatch
    channel.incoming <- frame
  }
  return nil
}