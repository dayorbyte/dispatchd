package main

import (
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) channelRoute(methodFrame amqp.MethodFrame) *AMQPError {
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
	var classId, methodId = methodFrame.MethodIdentifier()
	return NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) channelOpen(method *amqp.ChannelOpen) *AMQPError {
	if channel.state == CH_STATE_OPEN {
		var classId, methodId = method.MethodIdentifier()
		return NewHardError(504, "Channel already open", classId, methodId)
	}
	channel.sendMethod(&amqp.ChannelOpenOk{})
	channel.setStateOpen()
	return nil
}

func (channel *Channel) channelFlow(method *amqp.ChannelFlow) *AMQPError {
	channel.changeFlow(method.Active)
	channel.sendMethod(&amqp.ChannelFlowOk{channel.flow})
	return nil
}

func (channel *Channel) channelFlowOk(method *amqp.ChannelFlowOk) *AMQPError {
	var classId, methodId = method.MethodIdentifier()
	return NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) channelClose(method *amqp.ChannelClose) *AMQPError {
	// TODO(MAY): Report the class and method that are the reason for the close
	channel.sendMethod(&amqp.ChannelCloseOk{})
	channel.shutdown()
	return nil
}

func (channel *Channel) channelCloseOk(method *amqp.ChannelCloseOk) *AMQPError {
	channel.shutdown()
	return nil
}
