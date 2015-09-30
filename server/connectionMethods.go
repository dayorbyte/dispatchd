package main

import (
	"errors"
	"github.com/jeffjenkins/mq/amqp"
	"time"
)

func (channel *Channel) connectionRoute(methodFrame amqp.MethodFrame) error {
	switch method := methodFrame.(type) {
	case *amqp.ConnectionStartOk:
		return channel.connectionStartOk(method)
	case *amqp.ConnectionTuneOk:
		return channel.connectionTuneOk(method)
	case *amqp.ConnectionOpen:
		return channel.connectionOpen(method)
	case *amqp.ConnectionClose:
		return channel.connectionClose(method)
	case *amqp.ConnectionSecureOk:
		return channel.connectionSecureOk(method)
	case *amqp.ConnectionCloseOk:
		return channel.connectionCloseOk(method)
	case *amqp.ConnectionBlocked:
		return channel.connectionBlocked(method)
	case *amqp.ConnectionUnblocked:
		return channel.connectionUnblocked(method)
	}
	return errors.New("Unable to route method frame")
}

func (channel *Channel) connectionOpen(method *amqp.ConnectionOpen) error {
	// TODO(MAY): Add support for virtual hosts
	channel.conn.connectStatus.open = true
	channel.sendMethod(&amqp.ConnectionOpenOk{""})
	channel.conn.connectStatus.openOk = true
	return nil
}

func (channel *Channel) connectionTuneOk(method *amqp.ConnectionTuneOk) error {
	// TODO(MUST): Lower the limits from startOk if the client gives lower values
	// TODO(MUST): Start sending and monitoring heartbeats
	// TODO(MUST): If client gives higher frame max or channel max, hard close
	channel.conn.connectStatus.tuneOk = true
	if method.ChannelMax > channel.conn.maxChannels || method.FrameMax > channel.conn.maxFrameSize {
		channel.conn.hardClose()
		return nil
	}

	if method.Heartbeat > 0 {
		// Start sending heartbeats to the client
		channel.conn.sendHeartbeatInterval = time.Duration(method.Heartbeat) * time.Second
		channel.conn.handleSendHeartbeat()
	}
	// Start listening for heartbeats from the client
	channel.conn.handleClientHeartbeatTimeout()
	return nil
}

func (channel *Channel) connectionStartOk(method *amqp.ConnectionStartOk) error {
	// TODO(SHOULD): record product/version/platform/copyright/information
	// TODO(MUST): assert mechanism, response, locale are not null
	// TODO(MUST): if the auth is wrong, send 403 access-refused
	channel.conn.connectStatus.startOk = true

	if method.Mechanism != "PLAIN" {
		channel.conn.hardClose()
	}

	// TODO(MUST): add support these being enforced at the connection level.
	channel.sendMethod(&amqp.ConnectionTune{
		channel.conn.maxChannels,
		channel.conn.maxFrameSize,
		uint16(channel.conn.receiveHeartbeatInterval.Nanoseconds() / int64(time.Second)),
	})
	// TODO: Implement secure/secure-ok later if needed
	channel.conn.connectStatus.secure = true
	channel.conn.connectStatus.secureOk = true
	channel.conn.connectStatus.tune = true
	return nil
}

func (channel *Channel) startConnection() error {
	// TODO(SHOULD): add fields: host, product, version, platform, copyright, information
	channel.sendMethod(&amqp.ConnectionStart{0, 9, amqp.Table{}, []byte("PLAIN"), []byte("en_US")})
	return nil
}

func (channel *Channel) connectionClose(method *amqp.ConnectionClose) error {
	channel.sendMethod(&amqp.ConnectionCloseOk{})
	channel.conn.hardClose()
	return nil
}

func (channel *Channel) connectionSecureOk(method *amqp.ConnectionSecureOk) error {
	// TODO(MAY): If other security mechanisms are in place, handle this
	return errors.New("Server does not support secure/secure-ok. Use PLAIN auth in start-ok method")
}

func (channel *Channel) connectionCloseOk(method *amqp.ConnectionCloseOk) error {
	// TODO(MUST): Log class-id and method-id of the failing method, if available
	return nil
}

func (channel *Channel) connectionBlocked(method *amqp.ConnectionBlocked) error {
	// TODO(MUST): Error 540 NotImplemented
	return nil
}

func (channel *Channel) connectionUnblocked(method *amqp.ConnectionUnblocked) error {
	// TODO(MUST): Error 540 NotImplemented
	return nil
}
