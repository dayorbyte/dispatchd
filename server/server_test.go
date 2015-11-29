package main

import (
	"bytes"
	"encoding/json"
	"github.com/jeffjenkins/dispatchd/amqp"
	"github.com/jeffjenkins/dispatchd/util"
	"net"
	"testing"
)

func dbPath() string {
	return "/tmp/" + util.RandomId() + ".dispatchd.test.db"
}

func connFromServer(s *Server) *AMQPConnection {
	for _, conn := range s.conns {
		return conn
	}
	panic("no connections!")
}

func fromServerHelper(c net.Conn, fromServer chan *amqp.WireFrame) {
	// Reads bytes from a connection forever and throws them away
	// Useful for testing the internal server state rather than
	// the output sent to a client
	for {
		frame, err := amqp.ReadFrame(c)
		if err != nil {
			panic("Invalid frame from server! Check the wire protocol tests")
		}
		if frame.FrameType == 8 {
			// Skip heartbeats
			continue
		}
		fromServer <- frame
	}
}

func toServerHelper(c net.Conn, toServer chan *amqp.WireFrame) {
	for frame := range toServer {
		amqp.WriteFrame(c, frame)
	}
}

func methodToWireFrame(channelId uint16, method amqp.MethodFrame) *amqp.WireFrame {
	var buf = bytes.NewBuffer([]byte{})
	method.Write(buf)
	return &amqp.WireFrame{uint8(amqp.FrameMethod), channelId, buf.Bytes()}
}

func wireFrameToMethod(frame *amqp.WireFrame) amqp.MethodFrame {
	var methodReader = bytes.NewReader(frame.Payload)
	var methodFrame, err = amqp.ReadMethod(methodReader)
	if err != nil {
		panic("Failed to read method from server. Maybe not a method?")
	}
	return methodFrame
}

func logResponse(t *testing.T, fromServer chan *amqp.WireFrame) amqp.MethodFrame {
	var method = wireFrameToMethod(<-fromServer)
	var js, err = json.Marshal(method)
	if err != nil {
		panic("could not encode json, testing error!")
	}
	t.Logf(">>>> RECV '%s': %s", method.MethodName(), string(js))
	return method
}

func sendAndLogMethod(t *testing.T, channelId uint16, toServer chan *amqp.WireFrame, method amqp.MethodFrame) {
	var js, err = json.Marshal(method)
	if err != nil {
		panic("")
	}
	t.Logf("<<<< SENT '%s': %s", method.MethodName(), string(js))
	toServer <- methodToWireFrame(channelId, method)
}

func testServerHelper(t *testing.T, dbPath string, msgPath string) (s *Server, toServer chan *amqp.WireFrame, fromServer chan *amqp.WireFrame) {
	// Make channels
	// TODO: reduce these once we're reading/writng to the server
	toServer = make(chan *amqp.WireFrame, 500)
	fromServer = make(chan *amqp.WireFrame, 500)
	// Make server
	s = NewServer(dbPath, msgPath)
	s.init()

	// Make fake connection
	internal, external := net.Pipe()
	go fromServerHelper(external, fromServer)
	go toServerHelper(external, toServer)
	go s.openConnection(internal)
	// Set up connection
	external.Write([]byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1})
	logResponse(t, fromServer) // start
	sendAndLogMethod(t, 0, toServer, &amqp.ConnectionStartOk{
		ClientProperties: amqp.NewTable(),
		Mechanism:        "PLAIN",
		Response:         []byte("guest\x00guest"),
		Locale:           "en_US",
	})
	logResponse(t, fromServer) // tune
	sendAndLogMethod(t, 0, toServer, &amqp.ConnectionTuneOk{})
	sendAndLogMethod(t, 0, toServer, &amqp.ConnectionOpen{})
	logResponse(t, fromServer) // openok
	return
}
