package main

import (
	"github.com/jeffjenkins/dispatchd/util"
	"testing"
)

func TestTxCommitPublish(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	ch.Tx()
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	tc.wait(ch)
	if tc.s.queues["q1"].Len() != 0 {
		t.Fatalf("Tx failed to buffer messages")
	}
	ch.TxCommit()
	tc.wait(ch)
	if tc.s.queues["q1"].Len() != 3 {
		t.Fatalf("All messages were not added to queue")
	}
}

func TestTxRollbackPublish(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	ch.Tx()
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.TxRollback()
	ch.TxCommit()
	tc.wait(ch)
	if tc.s.queues["q1"].Len() != 0 {
		t.Fatalf("Tx Rollback still put messages in queue")
	}
}

func TestTxCommitAckNack(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	ch.Tx()
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.TxCommit()
	cTag := util.RandomId()
	deliveries, _ := ch.Consume("q1", cTag, false, false, false, false, NO_ARGS)
	<-deliveries
	<-deliveries
	lastMsg := <-deliveries
	lastMsg.Ack(true)
	tc.wait(ch)
	if len(tc.connFromServer().channels[1].awaitingAcks) != 3 {
		t.Fatalf("Acks were not held in tx")
	}
	ch.TxCommit()
	if len(tc.connFromServer().channels[1].awaitingAcks) != 0 {
		t.Fatalf("Acks were not processed on commit")
	}
}

func TestTxRollbackAckNack(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	ch.Tx()
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.TxCommit()

	cTag := util.RandomId()
	deliveries, _ := ch.Consume("q1", cTag, false, false, false, false, NO_ARGS)
	<-deliveries
	<-deliveries
	lastMsg := <-deliveries
	lastMsg.Nack(true, true)
	tc.wait(ch)
	if len(tc.connFromServer().channels[1].awaitingAcks) != 3 {
		t.Fatalf("Messages were acked despite rollback")
	}
}
