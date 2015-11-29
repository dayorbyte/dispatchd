package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"os"
	"testing"
)

func TestQueueMethods(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	s, toServer, fromServer := testServerHelper(t, path, msgPath)

	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
	})
	logResponse(t, fromServer)
	if len(s.queues) != 1 {
		t.Errorf("Wrong number of queues: %d", len(s.queues))
	}

	// Passive Check
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Durable:   true,
		Passive:   true,
	})
	logResponse(t, fromServer)

	// Bind
	sendAndLogMethod(t, chid, toServer, &amqp.QueueBind{
		Queue:      "q1",
		Exchange:   "amq.topic",
		RoutingKey: "rk.*.#",
		Arguments:  amqp.NewTable(),
	})
	logResponse(t, fromServer)
	if len(s.exchanges["amq.topic"].BindingsForQueue("q1")) != 1 {
		t.Errorf("Failed to bind to q1")
	}

	// Unbind
	sendAndLogMethod(t, chid, toServer, &amqp.QueueUnbind{
		Queue:      "q1",
		Exchange:   "amq.topic",
		RoutingKey: "rk.*.#",
		Arguments:  amqp.NewTable(),
	})
	logResponse(t, fromServer)
	if len(s.exchanges["amq.topic"].BindingsForQueue("q1")) != 0 {
		t.Errorf("Failed to unbind from q1")
	}

	// Delete
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDelete{
		Queue: "q1",
	})
	logResponse(t, fromServer)
	if len(s.queues) != 0 {
		t.Errorf("Wrong number of queues: %d", len(s.queues))
	}
}

func TestAutoAssignedQueue(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	_, toServer, fromServer := testServerHelper(t, path, msgPath)

	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "",
		Arguments: amqp.NewTable(),
	})
	resp := logResponse(t, fromServer).(*amqp.QueueDeclareOk)
	if len(resp.Queue) == 0 {
		t.Errorf("Autogenerate queue name failed")
	}
}

func TestBadQueueName(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	_, toServer, fromServer := testServerHelper(t, path, msgPath)

	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "~",
		Arguments: amqp.NewTable(),
	})
	resp := logResponse(t, fromServer).(*amqp.ChannelClose)
	if resp.ReplyCode != 406 {
		t.Errorf("Wrong response code")
	}
}

func TestPassiveNotFound(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	_, toServer, fromServer := testServerHelper(t, path, msgPath)

	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "does.not.exist",
		Arguments: amqp.NewTable(),
		Passive:   true,
	})
	resp := logResponse(t, fromServer).(*amqp.ChannelClose)
	if resp.ReplyCode != 404 {
		t.Errorf("Wrong response code")
	}
}

func TestExclusive(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	s, toServer, fromServer := testServerHelper(t, path, msgPath)

	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Exclusive: true,
	})
	logResponse(t, fromServer)

	// Check conn id
	conn := connFromServer(s)
	q, ok := s.queues["q1"]
	if !ok {
		t.Errorf("Could not find q1")
		return
	}
	if conn.id != q.ConnId {
		t.Errorf("Exclusive queue does not have connId set")
	}

	// Cheat and change the connection id so we don't need a second conn
	// for this test
	q.ConnId = 54321
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Exclusive: true,
	})

	resp := logResponse(t, fromServer).(*amqp.ChannelClose)
	if resp.ReplyCode != 405 {
		t.Errorf("Wrong response code")
	}
}

func TestNonMatchingQueue(t *testing.T) {
	// Setup
	path := dbPath()
	msgPath := dbPath()
	defer os.Remove(path)
	defer os.Remove(msgPath)
	_, toServer, fromServer := testServerHelper(t, path, msgPath)

	// Create channel
	var chid = uint16(1)
	sendAndLogMethod(t, chid, toServer, &amqp.ChannelOpen{})
	logResponse(t, fromServer)

	// Create Queue
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
	})
	logResponse(t, fromServer)

	// Create queue again with different args
	sendAndLogMethod(t, chid, toServer, &amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Durable:   true,
	})
	resp := logResponse(t, fromServer).(*amqp.ChannelClose)
	if resp.ReplyCode != 406 {
		t.Errorf("Wrong response code")
	}
}
