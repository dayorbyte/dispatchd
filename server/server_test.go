package main

import (
	"github.com/jeffjenkins/dispatchd/util"
	amqpclient "github.com/streadway/amqp"
	"net"
	"os"
	"testing"
	"time"
)

var NO_ARGS = make(amqpclient.Table)
var TEST_TRANSIENT_MSG = amqpclient.Publishing{
	Body: []byte("dispatchd"),
}

type testClient struct {
	t        *testing.T
	s        *Server
	serverDb string
	msgDb    string
}

func newTestClient(t *testing.T) *testClient {
	serverDb := dbPath()
	msgDb := dbPath()
	s := NewServer(serverDb, msgDb)
	s.init()
	tc := &testClient{
		t:        t,
		s:        s,
		serverDb: serverDb,
		msgDb:    msgDb,
	}
	return tc
}

func channelHelper(
	tc *testClient,
	conn *amqpclient.Connection,
) (
	*amqpclient.Channel,
	chan amqpclient.Return,
	chan *amqpclient.Error,
) {
	ch, err := conn.Channel()
	if err != nil {
		panic("Bad channel!")
	}
	retChan := make(chan amqpclient.Return)
	closeChan := make(chan *amqpclient.Error)
	ch.NotifyReturn(retChan)
	ch.NotifyClose(closeChan)
	return ch, retChan, closeChan
}

func (tc *testClient) connect() *amqpclient.Connection {
	internal, external := net.Pipe()
	go tc.s.openConnection(internal)
	// Set up connection
	clientconfig := amqpclient.Config{
		SASL:            nil,
		Vhost:           "/",
		ChannelMax:      100000,
		FrameSize:       100000,
		Heartbeat:       time.Duration(0),
		TLSClientConfig: nil,
		Properties:      make(amqpclient.Table),
		Dial: func(network, addr string) (net.Conn, error) {
			return external, nil
		},
	}

	client, err := amqpclient.DialConfig("amqp://localhost:1234", clientconfig)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func (tc *testClient) cleanup() {
	os.Remove(tc.msgDb)
	os.Remove(tc.serverDb)
	// tc.client.Close()
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
