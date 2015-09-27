
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

func (*Exchange) route(message interface{}) {

}