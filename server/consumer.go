package main

import (
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
)

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
	stopped       bool
}

func (consumer *Consumer) stop() {
	if !consumer.stopped {
		close(consumer.incoming)
		consumer.queue.removeConsumer(consumer.consumerTag)
	}
	consumer.stopped = true
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
	// fmt.Printf("+Active => %d\n", consumer.activeCount)
	consumer.channel.incrActive(size, bytes)
	consumer.limitLock.Lock()
	consumer.activeCount += size
	consumer.activeSize += bytes
	consumer.limitLock.Unlock()
	// fmt.Printf("+Active => %d\n", consumer.activeCount)
}

func (consumer *Consumer) decrActive(size uint16, bytes uint32) {
	// fmt.Printf("-Active => %d\n", consumer.activeCount)
	consumer.channel.decrActive(size, bytes)
	consumer.limitLock.Lock()
	consumer.activeCount -= size
	consumer.activeSize -= bytes
	consumer.limitLock.Unlock()
	// fmt.Printf("-Active => %d\n", consumer.activeCount)
}

func (consumer *Consumer) start() {
	for i := uint16(0); i < consumer.qos; i++ {
		go consumer.consume(i)
	}
}

func (consumer *Consumer) consume(id uint16) {
	fmt.Printf("Starting consumer %s#%d\n", consumer.consumerTag, id)
	for msg := range consumer.incoming {
		fmt.Printf("Consumer %s#%d got a message\n", consumer.consumerTag, id)
		var tag uint64 = 0
		if !consumer.noAck {
			tag = consumer.channel.addUnackedMessage(consumer, msg)
		}
		consumer.channel.sendContent(&amqp.BasicDeliver{
			ConsumerTag: consumer.consumerTag,
			DeliveryTag: tag,
			Redelivered: false,
			Exchange:    "",                  // TODO(MUST): the real exchange name
			RoutingKey:  consumer.queue.name, // TODO(must): real queue name
		}, msg)
	}
}
