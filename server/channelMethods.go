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
	// TODO(MUST): if channel is open, send 504 ChannelError
	channel.sendMethod(&amqp.ChannelOpenOk{})
	channel.state = CH_STATE_OPEN
	return nil
}

func (channel *Channel) channelFlow(method *amqp.ChannelFlow) error {
	// TODO(MUST): Error 540 NotImplemented
	return nil
}

func (channel *Channel) channelFlowOk(method *amqp.ChannelFlowOk) error {
	// TODO(MUST): Error 540 NotImplemented
	return nil
}

func (channel *Channel) channelClose(method *amqp.ChannelClose) error {
	// TODO(MAY): Report the class and method that are the reason for the close
	channel.sendMethod(&amqp.ChannelCloseOk{})
	channel.state = CH_STATE_CLOSED
	channel.destructor()
	return nil
}

func (channel *Channel) channelCloseOk(method *amqp.ChannelCloseOk) error {
	channel.state = CH_STATE_CLOSED
	channel.destructor()
	return nil
}
