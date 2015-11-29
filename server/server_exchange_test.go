package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"os"
	"testing"
)

func TestExchangeMethods(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	s, toServer, fromServer := testServerHelper(t, path, msgPath)
	if len(s.conns) != 1 {
		t.Errorf("Wrong number of open connections: %d", len(s.conns))
	}
	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create exchange
	sendAndLogMethod(t, chid, toServer, &amqp.ExchangeDeclare{
		Exchange:  "ex-1",
		Type:      "topic",
		Arguments: amqp.NewTable(),
	})
	logResponse(t, fromServer)
	if len(s.exchanges) != 5 {
		t.Errorf("Wrong number of exchanges: %d", len(s.exchanges))
	}

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q-1",
		Arguments: amqp.NewTable(),
	})
	logResponse(t, fromServer)
	if len(s.queues) != 1 {
		t.Errorf("Wrong number of queues: %d", len(s.queues))
	}

}
