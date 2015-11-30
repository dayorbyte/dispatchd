package main

import (
	"bytes"
	"encoding/json"
	"github.com/jeffjenkins/dispatchd/amqp"
	"github.com/jeffjenkins/dispatchd/util"
	"net"
	"os"
	"testing"
)

type testClient struct {
	t          *testing.T
	s          *Server
	toServer   chan *amqp.WireFrame
	fromServer chan *amqp.WireFrame
	intConn    net.Conn
	extConn    net.Conn
	serverDb   string
	msgDb      string
}

func newTestClient(t *testing.T) *testClient {
	serverDb := dbPath()
	msgDb := dbPath()
	s := NewServer(serverDb, msgDb)
	s.init()
	// Large buffers so we don't accidentally lock
	toServer := make(chan *amqp.WireFrame, 1000)
	fromServer := make(chan *amqp.WireFrame, 1000)

	// Make fake connection
	internal, external := net.Pipe()
	go s.openConnection(internal)
	// Set up connection
	external.Write([]byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1})
	tc := &testClient{
		t:          t,
		s:          s,
		toServer:   toServer,
		fromServer: fromServer,
		intConn:    internal,
		extConn:    external,
		serverDb:   serverDb,
		msgDb:      msgDb,
	}
	go tc.fromServerHelper()
	go tc.toServerHelper()

	tc.logResponse() // start
	tc.sendAndLogMethodWithChannel(0, &amqp.ConnectionStartOk{
		ClientProperties: amqp.NewTable(),
		Mechanism:        "PLAIN",
		Response:         []byte("guest\x00guest"),
		Locale:           "en_US",
	})
	tc.logResponse() // tune
	tc.sendAndLogMethodWithChannel(0, &amqp.ConnectionTuneOk{})
	tc.sendAndLogMethodWithChannel(0, &amqp.ConnectionOpen{})
	tc.logResponse() // openok
	// open channel
	tc.sendAndLogMethod(&amqp.ChannelOpen{})
	tc.logResponse()
	return tc
}

func (tc *testClient) cleanup() {
	os.Remove(tc.msgDb)
	os.Remove(tc.serverDb)
}

func dbPath() string {
	return "/tmp/" + util.RandomId() + ".dispatchd.test.db"
}

func (tc *testClient) connFromServer() *AMQPConnection {
	for _, conn := range tc.s.conns {
		return conn
	}
	panic("no connections!")
}

func (tc *testClient) fromServerHelper() {
	for {
		frame, err := amqp.ReadFrame(tc.extConn)
		if err != nil {
			panic("Invalid frame from server! Check the wire protocol tests")
		}
		if frame.FrameType == 8 {
			// Skip heartbeats
			continue
		}
		tc.fromServer <- frame
	}
}

func (tc *testClient) toServerHelper() {
	for frame := range tc.toServer {
		amqp.WriteFrame(tc.extConn, frame)
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

func (tc *testClient) logResponse() amqp.MethodFrame {
	var method = wireFrameToMethod(<-tc.fromServer)
	var js, err = json.Marshal(method)
	if err != nil {
		panic("could not encode json, testing error!")
	}
	tc.t.Logf(">>>> RECV '%s': %s", method.MethodName(), string(js))
	return method
}

func (tc *testClient) sendAndLogMethod(method amqp.MethodFrame) {
	var js, err = json.Marshal(method)
	if err != nil {
		panic("")
	}
	tc.t.Logf("<<<< SENT '%s': %s", method.MethodName(), string(js))
	tc.toServer <- methodToWireFrame(1, method)
}

func (tc *testClient) sendAndLogMethodWithChannel(channelId uint16, method amqp.MethodFrame) {
	var js, err = json.Marshal(method)
	if err != nil {
		panic("")
	}
	tc.t.Logf("<<<< SENT '%s': %s", method.MethodName(), string(js))
	tc.toServer <- methodToWireFrame(channelId, method)
}

func (tc *testClient) simplePublish(ex string, key string, msg string, durable bool) {
	// Send method
	tc.sendAndLogMethod(&amqp.BasicPublish{
		Exchange:   ex,
		RoutingKey: key,
	})
	// Send headers
	var buf = bytes.NewBuffer(make([]byte, 0))
	amqp.WriteShort(buf, uint16(60))
	amqp.WriteShort(buf, 0)
	amqp.WriteLonglong(buf, uint64(len(msg)))

	// TODO: write props. this sets all to not present
	amqp.WriteShort(buf, 0)

	tc.toServer <- &amqp.WireFrame{uint8(amqp.FrameHeader), 1, buf.Bytes()}
	// Send body
	tc.toServer <- &amqp.WireFrame{uint8(amqp.FrameBody), 1, []byte(msg)}
}
