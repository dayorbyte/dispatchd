package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"testing"
)

func TestExchangeMethods(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	// Create exchange
	tc.sendAndLogMethod(&amqp.ExchangeDeclare{
		Exchange:  "ex-1",
		Type:      "topic",
		Arguments: amqp.NewTable(),
	})
	tc.logResponse()
	if len(tc.s.exchanges) != 5 {
		t.Errorf("Wrong number of exchanges: %d", len(tc.s.exchanges))
	}

	// Create Queue
	tc.sendAndLogMethod(&amqp.QueueDeclare{
		Queue:     "q-1",
		Arguments: amqp.NewTable(),
	})
	tc.logResponse()
	if len(tc.s.queues) != 1 {
		t.Errorf("Wrong number of queues: %d", len(tc.s.queues))
	}

}
