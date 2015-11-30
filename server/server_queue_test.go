package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"testing"
)

func TestQueueMethods(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
	})
	tc.logResponse()
	if len(tc.s.queues) != 1 {
		t.Errorf("Wrong number of queues: %d", len(tc.s.queues))
	}

	// Passive Check
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Durable:   true,
		Passive:   true,
	})
	tc.logResponse()

	// Bind
	tc.sendAndLogMethod(&amqp.QueueBind{
		Queue:      "q1",
		Exchange:   "amq.topic",
		RoutingKey: "rk.*.#",
		Arguments:  amqp.NewTable(),
	})
	tc.logResponse()
	if len(tc.s.exchanges["amq.topic"].BindingsForQueue("q1")) != 1 {
		t.Errorf("Failed to bind to q1")
	}

	// Unbind
	tc.sendAndLogMethod(&amqp.QueueUnbind{
		Queue:      "q1",
		Exchange:   "amq.topic",
		RoutingKey: "rk.*.#",
		Arguments:  amqp.NewTable(),
	})
	tc.logResponse()
	if len(tc.s.exchanges["amq.topic"].BindingsForQueue("q1")) != 0 {
		t.Errorf("Failed to unbind from q1")
	}

	// Delete
	tc.sendAndLogMethod(&amqp.QueueDelete{
		Queue: "q1",
	})
	tc.logResponse()
	if len(tc.s.queues) != 0 {
		t.Errorf("Wrong number of queues: %d", len(tc.s.queues))
	}
}

func TestAutoAssignedQueue(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "",
		Arguments: amqp.NewTable(),
	})
	resp := tc.logResponse().(*amqp.QueueDeclareOk)
	if len(resp.Queue) == 0 {
		t.Errorf("Autogenerate queue name failed")
	}
}

func TestBadQueueName(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "~",
		Arguments: amqp.NewTable(),
	})
	resp := tc.logResponse().(*amqp.ChannelClose)
	if resp.ReplyCode != 406 {
		t.Errorf("Wrong response code")
	}
}

func TestPassiveNotFound(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "does.not.exist",
		Arguments: amqp.NewTable(),
		Passive:   true,
	})
	resp := tc.logResponse().(*amqp.ChannelClose)
	if resp.ReplyCode != 404 {
		t.Errorf("Wrong response code")
	}
}

func TestExclusive(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Exclusive: true,
	})
	tc.logResponse()

	// Check conn id
	conn := tc.connFromServer()
	q, ok := tc.s.queues["q1"]
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
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Exclusive: true,
	})

	resp := tc.logResponse().(*amqp.ChannelClose)
	if resp.ReplyCode != 405 {
		t.Errorf("Wrong response code")
	}
}

func TestNonMatchingQueue(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
	})
	tc.logResponse()

	// Create queue again with different args
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Durable:   true,
	})
	resp := tc.logResponse().(*amqp.ChannelClose)
	if resp.ReplyCode != 406 {
		t.Errorf("Wrong response code")
	}
}
