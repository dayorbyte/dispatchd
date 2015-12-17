package server

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"os"
	"runtime"
	"time"
)

func (channel *Channel) connectionRoute(conn *AMQPConnection, methodFrame amqp.MethodFrame) *amqp.AMQPError {
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
	var classId, methodId = methodFrame.MethodIdentifier()
	return amqp.NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) connectionOpen(conn *AMQPConnection, method *amqp.ConnectionOpen) *amqp.AMQPError {
	// TODO(MAY): Add support for virtual hosts. Check for access to the
	// selected one
	conn.connectStatus.open = true
	channel.SendMethod(&amqp.ConnectionOpenOk{""})
	conn.connectStatus.openOk = true
	return nil
}

func (channel *Channel) connectionTuneOk(conn *AMQPConnection, method *amqp.ConnectionTuneOk) *amqp.AMQPError {
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

func (channel *Channel) connectionStartOk(conn *AMQPConnection, method *amqp.ConnectionStartOk) *amqp.AMQPError {
	// TODO(SHOULD): record product/version/platform/copyright/information
	// TODO(MUST): assert mechanism, response, locale are not null
	conn.connectStatus.startOk = true

	if method.Mechanism != "PLAIN" {
		conn.hardClose()
	}

	if !conn.server.authenticate(method.Mechanism, method.Response) {
		var classId, methodId = method.MethodIdentifier()
		return &amqp.AMQPError{
			Code:   530,
			Class:  classId,
			Method: methodId,
			Msg:    "Authorization failed",
			Soft:   false,
		}
	}

	conn.clientProperties = method.ClientProperties
	// TODO(MUST): add support these being enforced at the connection level.
	channel.SendMethod(&amqp.ConnectionTune{
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

func (channel *Channel) startConnection() *amqp.AMQPError {
	// TODO(SHOULD): add fields: host, product, version, platform, copyright, information
	var capabilities = amqp.NewTable()
	capabilities.SetKey("publisher_confirms", false)
	capabilities.SetKey("basic.nack", true)
	var serverProps = amqp.NewTable()
	// TODO: the java rabbitmq client I'm using for load testing doesn't like these string
	//       fields even though the go/python clients do. If they are set as longstr (bytes)
	//       instead they work, so I'm doing that for now
	serverProps.SetKey("product", []byte("dispatchd"))
	serverProps.SetKey("version", []byte("0.1"))
	serverProps.SetKey("copyright", []byte("Jeffrey Jenkins, 2015"))
	serverProps.SetKey("capabilities", capabilities)
	serverProps.SetKey("platform", runtime.GOARCH)
	host, err := os.Hostname()
	if err != nil {
		serverProps.SetKey("host", []byte("UnknownHostError"))
	} else {
		serverProps.SetKey("host", []byte(host))
	}

	serverProps.SetKey("information", []byte("http://dispatchd.org"))

	channel.SendMethod(&amqp.ConnectionStart{0, 9, serverProps, []byte("PLAIN"), []byte("en_US")})
	return nil
}

func (channel *Channel) connectionClose(conn *AMQPConnection, method *amqp.ConnectionClose) *amqp.AMQPError {
	channel.SendMethod(&amqp.ConnectionCloseOk{})
	conn.hardClose()
	return nil
}

func (channel *Channel) connectionCloseOk(conn *AMQPConnection, method *amqp.ConnectionCloseOk) *amqp.AMQPError {
	conn.hardClose()
	return nil
}

func (channel *Channel) connectionSecureOk(conn *AMQPConnection, method *amqp.ConnectionSecureOk) *amqp.AMQPError {
	// TODO(MAY): If other security mechanisms are in place, handle this
	conn.hardClose()
	return nil
}

func (channel *Channel) connectionBlocked(conn *AMQPConnection, method *amqp.ConnectionBlocked) *amqp.AMQPError {
	return amqp.NewHardError(540, "Not implemented", 10, 60)
}

func (channel *Channel) connectionUnblocked(conn *AMQPConnection, method *amqp.ConnectionUnblocked) *amqp.AMQPError {
	return amqp.NewHardError(540, "Not implemented", 10, 61)
}
