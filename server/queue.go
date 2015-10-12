package main

import (
	"container/list"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
	"time"
)

type Message struct {
	header   *amqp.ContentHeaderFrame
	payload  []*amqp.WireFrame
	exchange string
	key      string
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
	consumers       *list.List // *Consumer
	currentConsumer *list.Element
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

func (q *Queue) add(message *Message) {
	// fmt.Printf("Queue \"%s\" got message! %d messages in the queue\n", q.name, q.queue.Len())
	// TODO: if there is a consumer available, dispatch
	if !q.closed {
		q.queueLock.Lock()
		defer q.queueLock.Unlock()
		q.queue.PushBack(message)
	}
}

func (q *Queue) readd(message *Message) {
	// TODO: if there is a consumer available, dispatch
	fmt.Println("Re-adding queue message!")
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	q.queue.PushFront(message)
}

func (q *Queue) removeConsumer(consumerTag string) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	fmt.Printf("Removing consumer %s\n", consumerTag)
	// reset current if needed
	if q.currentConsumer != nil && q.currentConsumer.Value.(*Consumer).consumerTag == consumerTag {
		q.currentConsumer = nil
	}
	// remove from list
	for e := q.consumers.Front(); e != nil; e = e.Next() {
		if e.Value.(*Consumer).consumerTag == consumerTag {
			fmt.Printf("Found consumer %s\n", consumerTag)
			q.consumers.Remove(e)
		}
	}
}

func (q *Queue) cancelConsumers() {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	// Send cancel to each consumer
	for c := q.consumers.Front(); c != nil; c = c.Next() {
		var consumer = c.Value.(*Consumer)
		consumer.channel.sendMethod(&amqp.BasicCancel{consumer.consumerTag, true})
	}
	// TODO: is it safe to delete while iterating over a list in go?
	toCancel := list.New()
	toCancel.PushBackList(q.consumers)
	for c := toCancel.Front(); c != nil; c = c.Next() {
		c.Value.(*Consumer).stop()
	}

	q.consumers.Init()

}

func (q *Queue) addConsumer(channel *Channel, method *amqp.BasicConsume) bool {
	fmt.Printf("Adding consumer\n")
	if q.closed {
		return false
	}
	var consumer = &Consumer{
		arguments:     method.Arguments,
		channel:       channel,
		consumerTag:   method.ConsumerTag,
		exclusive:     method.Exclusive,
		incoming:      make(chan *Message),
		noAck:         method.NoAck,
		noLocal:       method.NoLocal,
		qos:           1,
		queue:         q,
		prefetchSize:  channel.defaultPrefetchSize,
		prefetchCount: channel.defaultPrefetchCount,
	}
	channel.consumers[method.ConsumerTag] = consumer
	q.consumers.PushBack(consumer)
	consumer.start()
	return true
}

func (q *Queue) start() {
	fmt.Println("Queue started!")
	go func() {
		for {
			if q.closed {
				fmt.Printf("Queue closed!\n")
				break
			}
			// TODO: replace this check with a channel which notifies the queue
			// dispatch loop that someone is ready to accept more work.
			if q.queue.Len() == 0 || q.consumers.Len() == 0 {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			q.processOne()
		}
	}()
}

func (q *Queue) processOne() {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	if q.currentConsumer == nil {
		q.currentConsumer = q.consumers.Front()
	}
	var next = q.currentConsumer.Next()
	// fmt.Printf("Process 1\n")
	// Select the next round-robin consumer
	for {
		if next == q.currentConsumer { // full loop. nothing available
			break
		}
		if next == nil { // go back to start
			next = q.consumers.Front()
		}
		q.currentConsumer = next
		var consumer = next.Value.(*Consumer)
		if !consumer.ready() {
			// fmt.Printf("Consumer not ready!\n")
			continue
		}
		var msg = q.queue.Remove(q.queue.Front()).(*Message)

		if !consumer.noAck {
			// TODO(MUST): decr when we get acks
			consumer.incrActive(1, msg.size())
		}
		consumer.incoming <- msg
	}
}
