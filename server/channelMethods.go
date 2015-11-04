package main

import (
	"errors"
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
	if channel.state == CH_STATE_OPEN {
		var classId, methodId = method.MethodIdentifier()
		channel.conn.connectionErrorWithMethod(504, "Channel already open", classId, methodId)
		return nil
	}
	channel.sendMethod(&amqp.ChannelOpenOk{})
	channel.setStateOpen()
	return nil
}

func (channel *Channel) channelFlow(method *amqp.ChannelFlow) error {
	channel.changeFlow(method.Active)
	channel.sendMethod(&amqp.ChannelFlowOk{channel.flow})
	return nil
}

func (channel *Channel) channelFlowOk(method *amqp.ChannelFlowOk) error {
	var classId, methodId = method.MethodIdentifier()
	channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	return nil
}

func (channel *Channel) channelClose(method *amqp.ChannelClose) error {
	// TODO(MAY): Report the class and method that are the reason for the close
	channel.sendMethod(&amqp.ChannelCloseOk{})
	channel.shutdown()
	return nil
}

func (channel *Channel) channelCloseOk(method *amqp.ChannelCloseOk) error {
	channel.shutdown()
	return nil
}
