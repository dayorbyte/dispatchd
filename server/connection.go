package main

import (
	"fmt"
	"net"
	"time"
	// "errors"
	// "os"
	"bytes"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
)

// TODO: we can only be "in" one of these at once, so this should probably
// be one field
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
	id                       uint64
	nextChannel              int
	channels                 map[uint16]*Channel
	outgoing                 chan *amqp.WireFrame
	connectStatus            ConnectStatus
	server                   *Server
	network                  net.Conn
	lock                     sync.Mutex
	ttl                      time.Time
	sendHeartbeatInterval    time.Duration
	receiveHeartbeatInterval time.Duration
	maxChannels              uint16
	maxFrameSize             uint32
	clientProperties         *amqp.Table
}

func NewAMQPConnection(server *Server, network net.Conn) *AMQPConnection {
	return &AMQPConnection{
		id:            server.nextConnId(),
		network:       network,
		channels:      make(map[uint16]*Channel),
		outgoing:      make(chan *amqp.WireFrame),
		connectStatus: ConnectStatus{},
		server:        server,
		receiveHeartbeatInterval: 10 * time.Second,
		maxChannels:              4096,
		maxFrameSize:             65536,
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
	conn.connectStatus.closed = true
	conn.server.deregisterConnection(conn.id)
	for _, channel := range conn.channels {
		channel.shutdown()
	}
}

func (conn *AMQPConnection) setMaxChannels(max uint16) {
	conn.maxChannels = max
}

func (conn *AMQPConnection) setMaxFrameSize(max uint32) {
	conn.maxFrameSize = max
}

func (conn *AMQPConnection) startSendHeartbeat(interval time.Duration) {
	conn.sendHeartbeatInterval = interval
	conn.handleSendHeartbeat()
}

func (conn *AMQPConnection) handleSendHeartbeat() {
	go func() {
		for {
			if conn.connectStatus.closed {
				break
			}
			time.Sleep(conn.sendHeartbeatInterval / 2)
			conn.outgoing <- &amqp.WireFrame{8, 0, make([]byte, 0)}
		}
	}()
}

func (conn *AMQPConnection) handleClientHeartbeatTimeout() {
	// TODO(MUST): The spec is that any octet is a heartbeat substitute. Right
	// now this is only looking at frames, so a long send could cause a timeout
	// TODO(MUST): if the client isn't heartbeating how do we know when it's
	// gone?
	go func() {
		for {
			if conn.connectStatus.closed {
				break
			}
			time.Sleep(conn.receiveHeartbeatInterval / 2) //
			// If now is higher than TTL we need to time the client out
			if conn.ttl.Before(time.Now()) {
				conn.hardClose()
			}
		}
	}()
}

func (conn *AMQPConnection) handleOutgoing() {
	// TODO(MUST): Use SetWriteDeadline so we never wait too long. It should be
	// higher than the heartbeat in use. It should be reset after the heartbeat
	// interval is known.
	go func() {
		for {
			if conn.connectStatus.closed {
				break
			}
			var frame = <-conn.outgoing
			// fmt.Printf("Sending outgoing message. type: %d\n", frame.FrameType)
			// TODO(MUST): Hard close on irrecoverable errors, retry on recoverable
			// ones some number of times.

			amqp.WriteFrame(conn.network, frame)
			// for wire protocol debugging:
			// for _, b := range frame.Payload {
			// 	fmt.Printf("%d,", b)
			// }
			// fmt.Printf("\n")
		}
	}()
}

func (conn *AMQPConnection) connectionError(code uint16, message string) {
	conn.connectionErrorWithMethod(code, message, 0, 0)
}

func (conn *AMQPConnection) connectionErrorWithMethod(code uint16, message string, classId uint16, methodId uint16) {
	fmt.Println("Sending connection error:", message)
	conn.connectStatus.closing = true
	conn.channels[0].sendMethod(&amqp.ConnectionClose{code, message, classId, methodId})
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
			fmt.Println("Error reading frame: " + err.Error())
			conn.hardClose()
			break
		}

		// Upkeep. Remove things which have expired, etc
		conn.cleanUp()
		conn.ttl = time.Now().Add(conn.receiveHeartbeatInterval * 2)

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
