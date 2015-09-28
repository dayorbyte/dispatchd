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
		conn.destructor()
		return
	}

	var supported = []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
	if bytes.Compare(buf, supported) != 0 {
		conn.network.Write(supported)
		conn.destructor()
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

func (conn *AMQPConnection) destructor() {
	conn.network.Close()
	conn.server.deregisterConnection(conn.id)
}

func (conn *AMQPConnection) handleOutgoing() {
	go func() {
		for {
			var frame = <-conn.outgoing
			amqp.WriteFrame(conn.network, frame)
		}
	}()
}

func (conn *AMQPConnection) handleIncoming() {
	for {
		// If the connection is done, we stop handling frames
		if conn.connectStatus.closed {
			break
		}
		// Read from the network
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
			// heartbeat
			fmt.Println("Got heartbeat from client")
			continue
		}
		fmt.Printf("Got frame from client. Type: %d\n", frame.FrameType)
		// TODO: handle non-method frames (maybe?)

		// If we haven't finished the handshake, ignore frames on channels other
		// than 0
		if !conn.connectStatus.open && frame.Channel != 0 {
			fmt.Println("Non-0 channel for unopened connection")
			conn.network.Close()
			break
		}
		var channel, ok = conn.channels[frame.Channel]
		if !ok {
			channel = NewChannel(frame.Channel, conn)
			conn.channels[frame.Channel] = channel
			conn.channels[frame.Channel].start()
		}
		// Dispatch
		fmt.Println("Sending frame to", channel.id)
		channel.incoming <- frame
	}
}
