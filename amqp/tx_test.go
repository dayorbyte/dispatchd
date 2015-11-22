package amqp

import (
	"reflect"
	"testing"
)

func TestNewTxMessage(t *testing.T) {
	var msg = RandomMessage(false)
	var txMessage = NewTxMessage(msg, "hithere")
	if !reflect.DeepEqual(txMessage.GetMsg(), msg) {
		t.Errorf("Messages not equal")
	}
	if txMessage.GetQueueName() != "hithere" {
		t.Errorf("bad queue name")
	}
}

func TestNewTxAck(t *testing.T) {
	var tag uint64 = 12345 //, nack bool, requeueNack bool, multiple bool
	var txack = NewTxAck(tag, false, false, true)
	if txack.GetTag() != 12345 {
		t.Errorf("tag mismatch")
	}
	if !txack.Multiple {
		t.Errorf("bad multiple flag")
	}
	txack = NewTxAck(tag, false, true, false)
	if !txack.RequeueNack {
		t.Errorf("bad multiple requeue")
	}
	txack = NewTxAck(tag, true, false, true)
	if !txack.Nack {
		t.Errorf("bad multiple nack")
	}
}
