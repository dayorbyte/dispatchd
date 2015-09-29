package main

import (
	"fmt"
	"net"
	// "errors"
	// "os"
	"bytes"
	"github.com/jeffjenkins/mq/amqp"
)

type ConnectStatus struct {
	start    bool
	startOk  bool
	secure   bool
	secureOk bool
	tune     bool
	tuneOk   bool
	open     bool
	openOk   bool
	closing  bool
	closed   bool
}

type AMQPConnection struct {
	id            uint64
	nextChannel   int
	channels      map[uint16]*Channel
	outgoing      chan *amqp.WireFrame
	connectStatus ConnectStatus
	server        *Server
	network       net.Conn
}

func NewAMQPConnection(server *Server, network net.Conn) *AMQPConnection {
	return &AMQPConnection{
		id:            server.nextConnId(),
		network:       network,
		channels:      make(map[uint16]*Channel),
		outgoing:      make(chan *amqp.WireFrame),
		connectStatus: ConnectStatus{},
		server:        server,
	}
}

func (conn *AMQPConnection) openConnection() {
	// Negotiate Protocol
	buf := make([]byte, 8)
	_, err := conn.network.Read(buf)
	if err != nil {
		conn.hardClose()
		return
	}

	var supported = []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
	if bytes.Compare(buf, supported) != 0 {
		conn.network.Write(supported)
		conn.hardClose()
		return
	}

	// Create channel 0 and start the connection handshake
	conn.channels[0] = NewChannel(0, conn)
	conn.channels[0].start()
	conn.handleOutgoing()
	conn.handleIncoming()
}

func (conn *AMQPConnection) cleanUp() {

}

func (conn *AMQPConnection) deregisterChannel(id uint16) {
	delete(conn.channels, id)
}

func (conn *AMQPConnection) hardClose() {
	conn.network.Close()
	conn.server.deregisterConnection(conn.id)
	for _, channel := range conn.channels {
		if channel.state != CH_STATE_CLOSED {
			channel.destructor()
		}
	}
}

func (conn *AMQPConnection) handleOutgoing() {
	// TODO(MUST): Use SetWriteDeadline so we never wait too long. It should be
	// higher than the heartbeat in use. It should be reset after the heartbeat
	// interval is known.
	go func() {
		for {
			var frame = <-conn.outgoing
			// TODO(MUST): Hard close on irrecoverable errors, retry on recoverable
			// ones some number of times.
			amqp.WriteFrame(conn.network, frame)
		}
	}()
}

func (conn *AMQPConnection) connectionError(code uint16, message string) {
	// TODO(SHOULD): Add a timeout to hard close the connection if we get no reply
	conn.channels[0].sendMethod(&amqp.ConnectionClose{code, message, 0, 0})
}

func (conn *AMQPConnection) handleIncoming() {
	for {
		// If the connection is done, we stop handling frames
		if conn.connectStatus.closed {
			break
		}
		// Read from the network
		// TODO(MUST): Add a timeout to the read, esp. if there is no heartbeat
		// TODO(MUST): Hard close on unrecoverable errors, retry (with backoff?)
		// for recoverable ones
		frame, err := amqp.ReadFrame(conn.network)
		if err != nil {
			fmt.Println("Error reading frame")
			conn.network.Close()
			break
		}
		// Cleanup. Remove things which have expired, etc
		conn.cleanUp()

		switch {
		case frame.FrameType == 8:
			// TODO(MUST): Update last heartbeat time
			continue
		}

		if !conn.connectStatus.open && frame.Channel != 0 {
			fmt.Println("Non-0 channel for unopened connection")
			conn.hardClose()
			break
		}
		var channel, ok = conn.channels[frame.Channel]
		// TODO(MUST): Check that the channel number if in the valid range
		if !ok {
			channel = NewChannel(frame.Channel, conn)
			conn.channels[frame.Channel] = channel
			conn.channels[frame.Channel].start()
		}
		// Dispatch
		channel.incoming <- frame
	}
}
