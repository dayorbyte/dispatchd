package main

import (
	"github.com/jeffjenkins/dispatchd/util"
	"testing"
)

func TestAckNackOne(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)

	//
	// Publish and consume one message. Check that channel.awaitingAcks is updated
	// properly before and after acking the message
	//
	deliveries, err := ch.Consume("q1", "TestAckNackOne-1", false, false, false, false, NO_ARGS)
	if err != nil {
		t.Fatalf("Failed to consume")
	}

	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	msg := <-deliveries

	if len(tc.connFromServer().channels[1].awaitingAcks) != 1 {
		t.Fatalf("No awaiting ack for message just received")
	}

	msg.Ack(false)
	tc.wait(ch)
	if len(tc.connFromServer().channels[1].awaitingAcks) != 0 {
		t.Fatalf("No awaiting ack for message just received")
	}

	//
	// Publish and consume two messages. Check that awaiting acks is updated
	// and that nacking with and without the requeue options works
	//
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	msg1 := <-deliveries
	msg2 := <-deliveries

	tc.wait(ch)
	if len(tc.connFromServer().channels[1].awaitingAcks) != 2 {
		t.Fatalf("Should have 2 messages awaiting acks")
	}
	// Stop consuming so we can check requeue
	ch.Cancel("TestAckNackOne-1", false)
	msg1.Nack(false, false)
	msg2.Nack(false, true)
	tc.wait(ch)
	if tc.s.queues["q1"].Len() != 1 {
		t.Fatalf("Should have 1 message in queue")
	}

	deliveries, err = ch.Consume("q1", "TestAckNackOne-1", false, false, false, false, NO_ARGS)
	if err != nil {
		t.Fatalf("Failed to consume")
	}
	msg2_again := <-deliveries
	if !msg2_again.Redelivered {
		t.Fatalf("Redelivered message wasn't flagged")
	}
}

func TestAckNackMany(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	var consumerId = "TestAckNackMany-1"

	// Ack Many
	// Publish and consume two messages. Acck-multiple the second one and
	// check that both are acked
	//
	deliveries, err := ch.Consume("q1", consumerId, false, false, false, false, NO_ARGS)
	if err != nil {
		t.Fatalf("Failed to consume")
	}

	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	_ = <-deliveries
	msg2 := <-deliveries

	msg2.Ack(true)
	tc.wait(ch)
	if len(tc.connFromServer().channels[1].awaitingAcks) != 0 {
		t.Fatalf("No awaiting ack for message just received")
	}

	// Nack Many (no requeue)
	// Publish and consume two messages. Nack-multiple the second one and
	// check that both are acked and the queue is empty
	//
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	_ = <-deliveries
	msg2 = <-deliveries

	// Stop consuming so we can check requeue
	ch.Cancel(consumerId, false)
	msg2.Nack(true, false)
	tc.wait(ch)
	if tc.s.queues["q1"].Len() != 0 {
		t.Fatalf("Should have 0 message in queue")
	}

	// Nack Many (no requeue)
	// Publish and consume two messages. Nack-multiple the second one and
	// check that both are acked and the queue is empty
	//
	deliveries, err = ch.Consume("q1", consumerId, false, false, false, false, NO_ARGS)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	_ = <-deliveries
	msg2 = <-deliveries

	// Stop consuming so we can check requeue
	ch.Cancel(consumerId, false)
	msg2.Nack(true, true)
	tc.wait(ch)
	if tc.s.queues["q1"].Len() != 2 {
		t.Fatalf("Should have 2 message in queue")
	}
}

func TestRecover(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)
	cTag := util.RandomId()
	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)

	// Recover - no requeue
	// Publish two messages, consume them, call recover, consume them again
	//
	deliveries, err := ch.Consume("q1", cTag, false, false, false, false, NO_ARGS)
	if err != nil {
		t.Fatalf("Failed to consume")
	}

	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	tc.wait(ch)
	<-deliveries
	lastMsg := <-deliveries
	ch.Recover(false)
	<-deliveries
	<-deliveries
	lastMsg.Ack(true)

	// Recover - requeue
	// Publish two messages, consume them, call recover, check that they are
	// requeued
	//
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	tc.wait(ch)
	<-deliveries
	<-deliveries
	ch.Cancel(cTag, false)
	ch.Recover(true)
	tc.wait(ch)
	msgCount := tc.s.queues["q1"].Len()
	if msgCount != 2 {
		t.Fatalf("Should have 2 message in queue. Found", msgCount)
	}
}

func TestGet(t *testing.T) {
	//
	// Setup
	//
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)
	msg, ok, err := ch.Get("q1", false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !ok {
		t.Fatalf("Did not receive message")
	}
	if string(msg.Body) != string(TEST_TRANSIENT_MSG.Body) {
		t.Fatalf("wrong message response in get")
	}
}
