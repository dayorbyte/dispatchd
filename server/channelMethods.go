
package main

import (
  // "fmt"
  "errors"
  // "bytes"
  "github.com/jeffjenkins/mq/amqp"
)


func (channel *Channel) channelRoute(methodFrame amqp.MethodFrame) error {
  switch method := methodFrame.(type) {
  case *amqp.ChannelOpen:
    return channel.channelOpen(method)
  case *amqp.ChannelFlow:
    return channel.channelFlow(method)
  case *amqp.ChannelFlowOk:
    return channel.channelFlowOk(method)
  case *amqp.ChannelClose:
    return channel.channelClose(method)
  case *amqp.ChannelCloseOk:
    return channel.channelCloseOk(method)
  // case *amqp.ChannelOpenOk:
  //   return channel.channelOpenOk(method)
  }
  return errors.New("Unable to route method frame")
}


func (channel *Channel) channelOpen(method *amqp.ChannelOpen) error {
  channel.sendMethod(&amqp.ChannelOpenOk{})
  return nil
}

func (channel *Channel) channelFlow(method *amqp.ChannelFlow) error {
  // TODO
  return nil
}

func (channel *Channel) channelFlowOk(method *amqp.ChannelFlowOk) error {
  // TODO
  return nil
}

func (channel *Channel) channelClose(method *amqp.ChannelClose) error {
  // TODO: close channel
  channel.sendMethod(&amqp.ChannelCloseOk{})
  channel.state = CH_STATE_CLOSED
  return nil
}

func (channel *Channel) channelCloseOk(method *amqp.ChannelCloseOk) error {
  // TODO(close channel)
  channel.state = CH_STATE_CLOSED
  return nil
}
