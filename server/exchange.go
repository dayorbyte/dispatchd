
package main

import (
  "fmt"
  // "errors"
  // "bytes"
  "github.com/jeffjenkins/mq/amqp"
)

type Exchange struct {
  name string
  exType string
  passive bool
  durable bool
  autodelete bool
  internal bool
  nowait bool
  arguments amqp.Table
  incoming chan amqp.Frame
}

func (exchange *Exchange) start() {
  go func() {
    for {
      <- exchange.incoming
      fmt.Printf("Exchange %s got a frame\n", exchange.name)
    }
  }()
}

func (exchange *Exchange) publish(method *amqp.BasicPublish, header *amqp.ContentHeaderFrame, bodyFrames []*amqp.WireFrame) {
  fmt.Printf("Got message in exchange %s\n", exchange.name)
}