package main

import (
  "container/list"
	"fmt"
	"sync"
	"time"
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
	closed     bool
	arguments  amqp.Table
	queue      *list.List // *Message
	queueLock  sync.Mutex
	consumers  *list.List // *Consumer
	currentConsumer   *list.Element
}

func (q *Queue) add(message *Message) {
	fmt.Printf("Queue \"%s\" got message! %d messages in the queue\n", q.name, q.queue.Len())
	// TODO: if there is a consumer available, dispatch
	q.queue.PushBack(message)
}

func (q *Queue) addConsumer(channel *Channel, method *amqp.BasicConsume) {
	fmt.Printf("Adding consumer\n")
	var consumer = &Consumer{
		arguments: method.Arguments,
		channel: channel,
		consumerTag: method.ConsumerTag,
		exclusive: method.Exclusive,
		incoming: make(chan *Message),
		noAck: method.NoAck,
		noLocal: method.NoLocal,
		qos: 1,
		queue: q,
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
				var consumer = next.Value.(*Consumer)
				var msg = q.queue.Remove(q.queue.Front()).(*Message)
				consumer.incoming <- msg
				q.currentConsumer = next
			}
		}
	}()
}

type Consumer struct {
	arguments amqp.Table
	channel *Channel
	consumerTag string
	exclusive bool
	incoming chan *Message
	noAck bool
	noLocal bool
	qos uint16
	queue *Queue
}

func (consumer *Consumer) stop() {
	close(consumer.incoming)
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
			Exchange: "", // TODO(MUST): the real exchange name
			RoutingKey: consumer.queue.name, // TODO(must): real queue name
		}, msg)
	}
}