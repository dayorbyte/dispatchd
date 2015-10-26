package main

import (
	"encoding/json"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
)

type Consumer struct {
	arguments     amqp.Table
	channel       *Channel
	consumerTag   string
	exclusive     bool
	incoming      chan bool
	ackChan       chan bool
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
	statCount     uint64
}

func (consumer *Consumer) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"tag": consumer.consumerTag,
		"stats": map[string]interface{}{
			"total":             consumer.statCount,
			"qos":               consumer.qos,
			"active_size_bytes": consumer.activeSize,
			"active_count":      consumer.activeCount,
		},
		"ack": !consumer.noAck,
	})
}

func (consumer *Consumer) stop() {
	if !consumer.stopped {
		close(consumer.incoming)
		consumer.queue.removeConsumer(consumer.consumerTag)
	}
	consumer.stopped = true
}

func (consumer *Consumer) consumerReady() bool {
	if consumer.noAck {
		return true
	}
	consumer.limitLock.Lock()
	defer consumer.limitLock.Unlock()
	var sizeOk = consumer.prefetchSize == 0 || consumer.activeSize <= consumer.prefetchSize
	var bytesOk = consumer.prefetchCount == 0 || consumer.activeCount <= consumer.prefetchCount
	return sizeOk && bytesOk
}

func (consumer *Consumer) incrActive(size uint16, bytes uint32) {
	consumer.limitLock.Lock()
	consumer.activeCount += size
	consumer.activeSize += bytes
	consumer.limitLock.Unlock()
}

func (consumer *Consumer) decrActive(size uint16, bytes uint32) {
	consumer.limitLock.Lock()
	consumer.activeCount -= size
	consumer.activeSize -= bytes
	consumer.limitLock.Unlock()
}

func (consumer *Consumer) start() {
	fmt.Printf("Starting %d consumers\n", consumer.qos)
	for i := uint16(0); i < consumer.qos; i++ {
		go consumer.consume(i)
	}
}

func (consumer *Consumer) consume(id uint16) {
	fmt.Printf("[C:%s#%d]Starting consumer\n", consumer.consumerTag, id)
	consumer.queue.maybeReady <- false
	for _ = range consumer.incoming {
		// Check local limit
		if !consumer.consumerReady() {
			continue
		}
		// Try to get message/check channel limit
		var msg = consumer.queue.getOne(consumer.channel)
		if msg == nil {
			continue
		}
		var tag uint64 = 0
		if !consumer.noAck {
			tag = consumer.channel.addUnackedMessage(consumer, msg)
			consumer.incrActive(1, msg.size())
		}
		consumer.channel.sendContent(&amqp.BasicDeliver{
			ConsumerTag: consumer.consumerTag,
			DeliveryTag: tag,
			Redelivered: false,
			Exchange:    msg.exchange,
			RoutingKey:  msg.key,
		}, msg)
		consumer.statCount += 1
		// Wait on ack if applicable
		if !consumer.noAck {
			// TODO: this currently only deals with count QoS
			<-consumer.ackChan
		}
	}
}
