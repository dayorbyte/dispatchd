package server

import (
	"testing"
)

func TestQueueMethods(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	// Create Queue
	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)

	if len(tc.s.queues) != 1 {
		t.Errorf("Wrong number of queues: %d", len(tc.s.queues))
	}

	// Passive Check
	ch.QueueDeclarePassive("q1", true, false, false, false, NO_ARGS)

	// Bind
	ch.QueueBind("q1", "rk.*.#", "amq.topic", false, NO_ARGS)
	if len(tc.s.exchanges["amq.topic"].BindingsForQueue("q1")) != 1 {
		t.Errorf("Failed to bind to q1")
	}

	// Unbind
	ch.QueueUnbind("q1", "rk.*.#", "amq.topic", NO_ARGS)

	if len(tc.s.exchanges["amq.topic"].BindingsForQueue("q1")) != 0 {
		t.Errorf("Failed to unbind from q1")
	}

	// Delete
	ch.QueueDelete("q1", false, false, false)
	if len(tc.s.queues) != 0 {
		t.Errorf("Wrong number of queues: %d", len(tc.s.queues))
	}
}

func TestAutoAssignedQueue(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	// Create Queue
	resp, err := ch.QueueDeclare("", false, false, false, false, NO_ARGS)
	if err != nil {
		t.Fatalf("Error declaring queue")
	}
	if len(resp.Name) == 0 {
		t.Errorf("Autogenerate queue name failed")
	}
}

func TestBadQueueName(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, errChan := channelHelper(tc, conn)

	// Create Queue
	_, err := ch.QueueDeclare("!", false, false, false, true, NO_ARGS)
	if err != nil {
		panic("failed to declare queue")
	}
	resp := <-errChan
	if resp.Code != 406 {
		t.Errorf("Wrong response code")
	}
}

func TestPassiveNotFound(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, errChan := channelHelper(tc, conn)

	// Create Queue
	_, err := ch.QueueDeclarePassive("does.not.exist", true, false, false, true, NO_ARGS)
	if err != nil {
		panic("failed to declare queue")
	}
	resp := <-errChan
	if resp.Code != 404 {
		t.Errorf("Wrong response code")
	}
}

func TestPurge(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "a.b.c", "amq.topic", false, NO_ARGS)

	ch.Publish("amq.topic", "a.b.c", false, false, TEST_TRANSIENT_MSG)

	// This unbind is just to block us on the message being processed by the
	// channel so that the server has it.
	ch.QueueUnbind("q1", "a.b.c", "amq.topic", NO_ARGS)
	if tc.s.queues["q1"].Len() == 0 {
		t.Fatalf("Message did not make it into queue")
	}

	resp, err := ch.QueuePurge("q1", false)
	if err != nil {
		t.Fatalf("Failed to call QueuePurge")
	}

	if tc.s.queues["q1"].Len() != 0 {
		t.Fatalf("Message did not get purged from queue. Got %d", tc.s.queues["q1"].Len())
	}

	if resp != 1 {
		t.Fatalf("Purge did not return the right number of messages deleted")
	}
}

func TestExclusive(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, errChan := channelHelper(tc, conn)

	// Create Queue
	ch.QueueDeclare("q1", false, false, true, false, NO_ARGS)

	// Check conn id
	serverConn := tc.connFromServer()
	q, ok := tc.s.queues["q1"]
	if !ok {
		t.Fatalf("Could not find q1")
	}
	if serverConn.id != q.ConnId {
		t.Fatalf("Exclusive queue does not have connId set")
	}

	// Cheat and change the connection id so we don't need a second conn
	// for this test
	q.ConnId = 54321
	// NOTE: if nowait isn't true this blocks forever
	ch.QueueDeclare("q1", false, false, true, true, NO_ARGS)

	resp := <-errChan
	if resp.Code != 405 {
		t.Errorf("Wrong response code")
	}
}

func TestNonMatchingQueue(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, errChan := channelHelper(tc, conn)

	// Create Queue
	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)

	// Create queue again with different args
	// NOTE: if nowait isn't true this blocks forever
	ch.QueueDeclare("q1", true, false, false, true, NO_ARGS)
	resp := <-errChan
	if resp.Code != 406 {
		t.Errorf("Wrong response code")
	}
}
