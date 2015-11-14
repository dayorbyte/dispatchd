package amqp

import (
	// "fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var counter int64

func init() {
	rand.Seed(time.Now().UnixNano())
	counter = 0
}

func nextId() int64 {
	return atomic.AddInt64(&counter, 1)
}

func RandomMessage(persistent bool) *Message {
	var size, payload = RandomPayload()
	var randomHeader = RandomHeader(size, persistent)
	var exchange = "exchange-name"
	var routingKey = "routing-key"
	var randomPublish = &BasicPublish{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Mandatory:  false,
		Immediate:  false,
	}
	return &Message{
		Id:          nextId(),
		Header:      randomHeader,
		Payload:     payload,
		Exchange:    exchange,
		Key:         routingKey,
		Method:      randomPublish,
		Redelivered: 0,
		LocalId:     nextId(),
	}
}

func RandomPayload() (uint64, []*WireFrame) {
	var size uint64 = 500
	var payload = make([]byte, size)
	return size, []*WireFrame{
		&WireFrame{
			FrameType: 3,
			Channel:   1,
			Payload:   payload,
		},
	}
}

func RandomHeader(size uint64, persistent bool) *ContentHeaderFrame {
	var props = BasicContentHeaderProperties{}
	if persistent {
		var b = byte(2)
		props.DeliveryMode = &b
	} else {
		var b = byte(1)
		props.DeliveryMode = &b
	}
	return &ContentHeaderFrame{
		ContentClass:    60,
		ContentWeight:   0,
		ContentBodySize: size,
		PropertyFlags:   0,
		Properties:      &props,
	}
}
