package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"testing"
	"time"
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

func TestPurge(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q1",
		Arguments: amqp.NewTable(),
		Exclusive: true,
	})
	tc.logResponse()

	tc.sendAndLogMethod(&amqp.QueueBind{
		Exchange:   "amq.topic",
		Queue:      "q1",
		RoutingKey: "a.b.c",
		Arguments:  amqp.NewTable(),
	})
	tc.logResponse()

	tc.simplePublish("amq.topic", "a.b.c", "hello world")

	// TODO: handle this more elegantly than a sleep. Best bet is probably sending
	// a random method and waiting for the response since messages are processed
	// in order
	time.Sleep(5 * time.Millisecond)
	if tc.s.queues["q1"].Len() == 0 {
		t.Fatalf("Message did not make it into queue")
	}

	tc.sendAndLogMethod(&amqp.QueuePurge{
		Queue: "q1",
	})
	resp := tc.logResponse().(*amqp.QueuePurgeOk)

	if tc.s.queues["q1"].Len() != 0 {
		t.Fatalf("Message did not get purged from queue. Got %d", tc.s.queues["q1"].Len())
	}

	if resp.MessageCount != 1 {
		t.Fatalf("Purge did not return the right number of messages deleted")
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
