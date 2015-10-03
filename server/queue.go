package main

import (
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
)

type Message struct {
	header  *amqp.ContentHeaderFrame
	payload []*amqp.WireFrame
}

type Queue struct {
	name       string
	durable    bool
	exclusive  bool
	autoDelete bool
	arguments  amqp.Table
	queue      chan *Message
}

func (q *Queue) add(message *Message) {
	fmt.Printf("Queue \"%s\" got message!\n", q.name)
	q.queue <- message
}
