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

func (q *Queue) add(message *Message) {
	fmt.Printf("Queue \"%s\" got message! %d messages in the queue\n", q.name, q.queue.Len())
	// TODO: if there is a consumer available, dispatch
	q.queue.PushBack(message)
}

func (q *Queue) addConsumer(channel *Channel, method *amqp.BasicConsume) {
	fmt.Printf("Adding consumer\n")
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
	return

}

func (q *Queue) start() {
	go func() {
		for {
			if q.closed {
				break
			}
			if q.queue.Len() == 0 || q.consumers.Len() == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if q.currentConsumer == nil {
				q.currentConsumer = q.consumers.Front()
			}
			var next = q.currentConsumer.Next()

			// Select the next round-robin consumer
			for {
				if next == q.currentConsumer { // full loop
					break
				}
				if next == nil { // go back to start
					next = q.consumers.Front()
				}
				q.currentConsumer = next
				var consumer = next.Value.(*Consumer)
				if !consumer.ready() {
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
	}()
}

type Consumer struct {
	arguments     amqp.Table
	channel       *Channel
	consumerTag   string
	exclusive     bool
	incoming      chan *Message
	noAck         bool
	noLocal       bool
	qos           uint16
	queue         *Queue
	limitLock     sync.Mutex
	prefetchSize  uint32
	prefetchCount uint16
	activeSize    uint32
	activeCount   uint16
}

func (consumer *Consumer) stop() {
	close(consumer.incoming)
}

func (consumer *Consumer) ready() bool {
	if consumer.noAck {
		return true
	}
	if !consumer.channel.consumeLimitsOk() {
		return false
	}
	var sizeOk = consumer.prefetchSize == 0 || consumer.activeSize <= consumer.prefetchSize
	var bytesOk = consumer.prefetchCount == 0 || consumer.activeCount <= consumer.prefetchCount
	return sizeOk && bytesOk
}

func (consumer *Consumer) incrActive(size uint16, bytes uint32) {
	consumer.channel.incrActive(size, bytes)
	consumer.limitLock.Lock()
	consumer.activeCount += size
	consumer.activeSize += bytes
	consumer.limitLock.Unlock()
}

func (consumer *Consumer) decrActive(size uint16, bytes uint32) {
	consumer.channel.incrActive(size, bytes)
	consumer.limitLock.Lock()
	consumer.activeCount -= size
	consumer.activeSize -= bytes
	consumer.limitLock.Unlock()
}

func (consumer *Consumer) start() {
	for i := uint16(0); i < consumer.qos; i++ {
		go consumer.consume(i)
	}
}

func (consumer *Consumer) consume(id uint16) {
	var tag = uint64(0)
	fmt.Printf("Starting consumer %s#%d\n", consumer.consumerTag, id)
	for msg := range consumer.incoming {
		// TODO(MUST): stop if channel is closed
		if consumer.qos < id {
			break
		}
		tag++
		fmt.Printf("Consumer %s#%d got a message\n", consumer.consumerTag, id)
		consumer.channel.sendContent(&amqp.BasicDeliver{
			ConsumerTag: consumer.consumerTag,
			DeliveryTag: tag,
			Redelivered: false,
			Exchange:    "",                  // TODO(MUST): the real exchange name
			RoutingKey:  consumer.queue.name, // TODO(must): real queue name
		}, msg)
	}
}
