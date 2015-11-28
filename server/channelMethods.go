package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
)

func (channel *Channel) channelRoute(methodFrame amqp.MethodFrame) *amqp.AMQPError {
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
	return amqp.NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) channelOpen(method *amqp.ChannelOpen) *amqp.AMQPError {
	if channel.state == CH_STATE_OPEN {
		var classId, methodId = method.MethodIdentifier()
		return amqp.NewHardError(504, "Channel already open", classId, methodId)
	}
	channel.SendMethod(&amqp.ChannelOpenOk{})
	channel.setStateOpen()
	return nil
}

func (channel *Channel) channelFlow(method *amqp.ChannelFlow) *amqp.AMQPError {
	channel.changeFlow(method.Active)
	channel.SendMethod(&amqp.ChannelFlowOk{channel.flow})
	return nil
}

func (channel *Channel) channelFlowOk(method *amqp.ChannelFlowOk) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	return amqp.NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) channelClose(method *amqp.ChannelClose) *amqp.AMQPError {
	// TODO(MAY): Report the class and method that are the reason for the close
	channel.SendMethod(&amqp.ChannelCloseOk{})
	channel.shutdown()
	return nil
}

func (channel *Channel) channelCloseOk(method *amqp.ChannelCloseOk) *amqp.AMQPError {
	channel.shutdown()
	return nil
}
