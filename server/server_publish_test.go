package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
	"testing"
)

func TestImmediate(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	tc.declareQueue("q1")
	tc.bindQueue("amq.direct", "q1", "abc")

	tc.sendAndLogMethod(&amqp.BasicPublish{
		Exchange:   "amq.direct",
		RoutingKey: "abc",
		Immediate:  true,
	})
	var msg = "dispatchd"
	tc.sendSimpleContentHeader(msg)
	tc.sendMessageFrames(msg)

	var method, _, body = tc.logReturn1()
	if method.(*amqp.BasicReturn).ReplyCode != 313 {
		t.Fatalf("Wrong reply code with Immediate return")
	}
	if string(body.Payload) != msg {
		t.Fatalf("Did not get same payload back in BasicReturn")
	}
}

func TestMandatory(t *testing.T) {
	tc := newTestClient(t)
	defer tc.cleanup()

	tc.sendAndLogMethod(&amqp.BasicPublish{
		Exchange:   "amq.direct",
		RoutingKey: "abc",
		Mandatory:  true,
	})
	var msg = "dispatchd"
	tc.sendSimpleContentHeader(msg)
	tc.sendMessageFrames(msg)

	var method, _, body = tc.logReturn1()
	if method.(*amqp.BasicReturn).ReplyCode != 313 {
		t.Fatalf("Wrong reply code with Immediate return")
	}
	if string(body.Payload) != msg {
		t.Fatalf("Did not get same payload back in BasicReturn")
	}
}
