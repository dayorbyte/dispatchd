package main

import (
  "fmt"
  "github.com/jeffjenkins/mq/amqp"
)

type Channel struct {
  incoming chan *amqp.FrameWrapper
}

func (channel *Channel) start() {
  go func() {
    for {
      channel.handleFrame(<- channel.incoming)
    }
  }()
}

func (channel *Channel) handleFrame(frame *amqp.FrameWrapper) {
  fmt.Println("Got frame")
}