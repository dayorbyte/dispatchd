package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
	// "time"
)

type Message struct {
	header   *amqp.ContentHeaderFrame
	payload  []*amqp.WireFrame
	exchange string
	key      string
	method   *amqp.BasicPublish
}

func NewMessage(method *amqp.BasicPublish) *Message {
	return &Message{
		method:   method,
		exchange: method.Exchange,
		key:      method.RoutingKey,
		payload:  make([]*amqp.WireFrame, 0, 1),
	}
}

type UnackedMessage struct {
	consumer *Consumer
	msg      *Message // what's the message?
	queue    *Queue   // where do we return this on failure?
}

func (message *Message) size() uint32 {
	// TODO: include header size
	var size uint32 = 0
	for _, frame := range message.payload {
		size += uint32(len(frame.Payload))
	}
	return size
}

type Queue struct {
	name            string
	durable         bool
	exclusive       bool
	autoDelete      bool
	closed          bool
	arguments       amqp.Table
	queue           *list.List // *Message
	queueLock       sync.Mutex
	consumerLock    sync.RWMutex
	consumers       []*Consumer // *Consumer
	currentConsumer int
	statCount       uint64
	maybeReady      chan bool
}

func (q *Queue) nextConsumer() *Consumer {
	// TODO: also check availability
	q.consumerLock.RLock()
	defer q.consumerLock.RUnlock()
	var num = len(q.consumers)
	if num == 0 {
		return nil
	}
	q.currentConsumer = (q.currentConsumer + 1) % len(q.consumers)
	return q.consumers[q.currentConsumer]
}

func (q *Queue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":       q.name,
		"durable":    q.durable,
		"exclusive":  q.exclusive,
		"autoDelete": q.autoDelete,
		"size":       q.queue.Len(),
		"consumers":  q.consumers,
	})
}

func (q *Queue) close() {
	// This discards any messages which would be added. It does not
	// do cleanup
	q.closed = true
}

func (q *Queue) purge() uint32 {
	var length = q.queue.Len()
	q.queue.Init()
	return uint32(length)
}

func (q *Queue) add(channel *Channel, message *Message) {
	// TODO: if there is a consumer available, dispatch
	if message.method.Immediate {
		panic("Immediate not implemented!")
		// TODO: deliver this message somewhere if possible, otherwise:
		channel.sendContent(&amqp.BasicReturn{
			ReplyCode: 313, // TODO: what code?
			ReplyText: "No consumers available",
		}, message)
	}

	if !q.closed {
		q.queueLock.Lock()
		q.statCount += 1
		q.queue.PushBack(message)
		q.queueLock.Unlock()
		select {
		case q.maybeReady <- true:
		}
	}
}

func (q *Queue) readd(message *Message) {
	// TODO: if there is a consumer available, dispatch
	fmt.Println("Re-adding queue message!")
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	q.queue.PushFront(message)
	q.maybeReady <- true
}

func (q *Queue) removeConsumer(consumerTag string) {
	q.consumerLock.Lock()
	defer q.consumerLock.Unlock()

	// remove from list
	for i, c := range q.consumers {
		if c.consumerTag == consumerTag {
			fmt.Printf("Found consumer %s\n", consumerTag)
			q.consumers = append(q.consumers[:i], q.consumers[i+1:]...)
		}
	}
	var size = len(q.consumers)
	if size == 0 {
		q.currentConsumer = 0
	} else {
		q.currentConsumer = q.currentConsumer % size
	}

}

func (q *Queue) cancelConsumers() {
	q.consumerLock.Lock()
	defer q.consumerLock.Unlock()
	// Send cancel to each consumer
	for _, c := range q.consumers {
		c.channel.sendMethod(&amqp.BasicCancel{c.consumerTag, true})
		c.stop()
	}
	q.consumers = make([]*Consumer, 0, 1)
}

func (q *Queue) addConsumer(channel *Channel, method *amqp.BasicConsume) bool {
	fmt.Printf("Adding consumer\n")
	if q.closed {
		return false
	}
	var consumer = &Consumer{
		arguments:   method.Arguments,
		channel:     channel,
		consumerTag: method.ConsumerTag,
		exclusive:   method.Exclusive,
		incoming:    make(chan bool),
		ackChan:     make(chan bool),
		noAck:       method.NoAck,
		noLocal:     method.NoLocal,
		// TODO: size QOS
		qos:           channel.defaultPrefetchCount,
		queue:         q,
		prefetchSize:  channel.defaultPrefetchSize,
		prefetchCount: channel.defaultPrefetchCount,
	}
	q.consumerLock.Lock()
	channel.addConsumer(consumer)
	q.consumers = append(q.consumers, consumer)
	q.consumerLock.Unlock()
	consumer.start()
	return true
}

func (q *Queue) start() {
	fmt.Println("Queue started!")
	go func() {
		for _ = range q.maybeReady {
			if q.closed {
				fmt.Printf("Queue closed!\n")
				break
			}
			q.processOne()
		}
	}()
}

func (q *Queue) processOne() {
	q.consumerLock.RLock()
	defer q.consumerLock.RUnlock()
	var size = len(q.consumers)
	if size == 0 {
		return
	}
	for count := 0; count < size; count++ {
		q.currentConsumer = (q.currentConsumer + 1) % size
		var c = q.consumers[q.currentConsumer]
		select {
		case c.incoming <- true:
			break
		default:
		}

	}
}

func (q *Queue) getOne() *Message {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	if q.queue.Len() == 0 {
		return nil
	}
	return q.queue.Remove(q.queue.Front()).(*Message)
}
