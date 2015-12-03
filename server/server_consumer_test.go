package main

import (
	"testing"
)

func TestAckNack(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	ch, _, _ := channelHelper(tc, conn)

	ch.QueueDeclare("q1", false, false, false, false, NO_ARGS)
	ch.QueueBind("q1", "abc", "amq.direct", false, NO_ARGS)
	ch.Publish("amq.direct", "abc", false, false, TEST_TRANSIENT_MSG)

	deliveries, err := ch.Consume(
		"q1",
		"TestAckNack",
		false,
		false,
		false,
		false,
		NO_ARGS,
	)
	if err != nil {
		t.Fatalf("Failed to consume")
	}
	msg1 := <-deliveries

	if len(tc.connFromServer().channels[1].awaitingAcks) != 1 {
		t.Fatalf("No awaiting ack for message just received")
	}

	msg1.Ack(false)
	ch.QueueUnbind("q1", "abc", "amq.direct", NO_ARGS)
	if len(tc.connFromServer().channels[1].awaitingAcks) != 0 {
		t.Fatalf("No awaiting ack for message just received")
	}
}
