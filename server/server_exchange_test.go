package main

import (
	"testing"
)

func TestExchangeMethods(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()
	conn := tc.connect()
	channel, _, _ := channelHelper(tc, conn)

	channel.ExchangeDeclare("ex-1", "topic", false, false, false, false, NO_ARGS)

	// Create exchange
	if len(tc.s.exchanges) != 5 {
		t.Errorf("Wrong number of exchanges: %d", len(tc.s.exchanges))
	}

	// Create Queue
	channel.QueueDeclare("q-1", false, false, false, false, NO_ARGS)
	if len(tc.s.queues) != 1 {
		t.Errorf("Wrong number of queues: %d", len(tc.s.queues))
	}

	// Delete exchange
	channel.ExchangeDelete("ex-1", false, false)
	if len(tc.s.exchanges) != 4 {
		t.Errorf("Wrong number of exchanges: %d", len(tc.s.exchanges))
	}
}
