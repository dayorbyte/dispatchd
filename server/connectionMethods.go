package main

import (
	"errors"
	"github.com/jeffjenkins/mq/amqp"
	"time"
)

func (channel *Channel) connectionRoute(conn *AMQPConnection, methodFrame amqp.MethodFrame) error {
	switch method := methodFrame.(type) {
	case *amqp.ConnectionStartOk:
		return channel.connectionStartOk(conn, method)
	case *amqp.ConnectionTuneOk:
		return channel.connectionTuneOk(conn, method)
	case *amqp.ConnectionOpen:
		return channel.connectionOpen(conn, method)
	case *amqp.ConnectionClose:
		return channel.connectionClose(conn, method)
	case *amqp.ConnectionSecureOk:
		return channel.connectionSecureOk(conn, method)
	case *amqp.ConnectionCloseOk:
		return channel.connectionCloseOk(conn, method)
	case *amqp.ConnectionBlocked:
		return channel.connectionBlocked(conn, method)
	case *amqp.ConnectionUnblocked:
		return channel.connectionUnblocked(conn, method)
	}
	return errors.New("Unable to route method frame")
}

func (channel *Channel) connectionOpen(conn *AMQPConnection, method *amqp.ConnectionOpen) error {
	// TODO(MAY): Add support for virtual hosts. Check for access to the
	// selected one
	conn.connectStatus.open = true
	channel.sendMethod(&amqp.ConnectionOpenOk{""})
	conn.connectStatus.openOk = true
	return nil
}

func (channel *Channel) connectionTuneOk(conn *AMQPConnection, method *amqp.ConnectionTuneOk) error {
	conn.connectStatus.tuneOk = true
	if method.ChannelMax > conn.maxChannels || method.FrameMax > conn.maxFrameSize {
		conn.hardClose()
		return nil
	}

	conn.setMaxChannels(method.ChannelMax)
	conn.setMaxFrameSize(method.FrameMax)

	if method.Heartbeat > 0 {
		// Start sending heartbeats to the client
		conn.startSendHeartbeat(time.Duration(method.Heartbeat) * time.Second)
	}
	// Start listening for heartbeats from the client.
	// We always ask for them since we want to shut down
	// connections not in use
	conn.handleClientHeartbeatTimeout()
	return nil
}

func (channel *Channel) connectionStartOk(conn *AMQPConnection, method *amqp.ConnectionStartOk) error {
	// TODO(SHOULD): record product/version/platform/copyright/information
	// TODO(MUST): assert mechanism, response, locale are not null
	// TODO(MUST): if the auth is wrong, send 403 access-refused
	conn.connectStatus.startOk = true

	if method.Mechanism != "PLAIN" {
		conn.hardClose()
	}

	// TODO(MUST): add support these being enforced at the connection level.
	channel.sendMethod(&amqp.ConnectionTune{
		conn.maxChannels,
		conn.maxFrameSize,
		uint16(conn.receiveHeartbeatInterval.Nanoseconds() / int64(time.Second)),
	})
	// TODO: Implement secure/secure-ok later if needed
	conn.connectStatus.secure = true
	conn.connectStatus.secureOk = true
	conn.connectStatus.tune = true
	return nil
}

func (channel *Channel) startConnection() error {
	// TODO(SHOULD): add fields: host, product, version, platform, copyright, information
	var capabilities = make(amqp.Table)
	capabilities["publisher_confirms"] = true
	capabilities["basic.nack"] = true
	var serverProps = make(amqp.Table)
	serverProps["capabilities"] = capabilities
	// TODO: the java rabbitmq client I'm using for load testing doesn't like these string
	//       fields even though the go/python clients do. commenting out until I can
	//       compile and get some debugging info out of that client
	// serverProps["product"] = "mq"
	// serverProps["version"] = "0.1"
	// serverProps["copyright"] = "Jeffrey Jenkins, 2015"
	// serverProps["platform"] = "TODO"
	// serverProps["host"] = "TODO"
	// serverProps["information"] = "https://github.com/jeffjenkins/mq"

	channel.sendMethod(&amqp.ConnectionStart{0, 9, serverProps, []byte("PLAIN"), []byte("en_US")})
	return nil
}

func (channel *Channel) connectionClose(conn *AMQPConnection, method *amqp.ConnectionClose) error {
	channel.sendMethod(&amqp.ConnectionCloseOk{})
	conn.hardClose()
	return nil
}

func (channel *Channel) connectionCloseOk(conn *AMQPConnection, method *amqp.ConnectionCloseOk) error {
	conn.hardClose()
	return nil
}

func (channel *Channel) connectionSecureOk(conn *AMQPConnection, method *amqp.ConnectionSecureOk) error {
	// TODO(MAY): If other security mechanisms are in place, handle this
	conn.hardClose()
	return nil
}

func (channel *Channel) connectionBlocked(conn *AMQPConnection, method *amqp.ConnectionBlocked) error {
	conn.connectionErrorWithMethod(540, "Not implemented", 10, 60)
	return nil
}

func (channel *Channel) connectionUnblocked(conn *AMQPConnection, method *amqp.ConnectionUnblocked) error {
	conn.connectionErrorWithMethod(540, "Not implemented", 10, 61)
	return nil
}
