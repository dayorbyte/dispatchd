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
	noAck         bool
	noLocal       bool
	qos           uint16
	queue         *Queue
	consumeLock   sync.Mutex
	limitLock     sync.Mutex
	prefetchSize  uint32
	prefetchCount uint16
	activeSize    uint32
	activeCount   uint16
	stopLock      sync.Mutex
	stopped       bool
	statCount     uint64
	localId       int64
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
		consumer.stopLock.Lock()
		consumer.stopped = true
		close(consumer.incoming)
		consumer.stopLock.Unlock()
		consumer.queue.removeConsumer(consumer.consumerTag)
	}
}

func (consumer *Consumer) consumerReady() bool {
	if !consumer.channel.flow {
		return false
	}
	if consumer.noAck {
		return true
	}
	consumer.limitLock.Lock()
	defer consumer.limitLock.Unlock()
	var sizeOk = consumer.prefetchSize == 0 || consumer.activeSize < consumer.prefetchSize
	var bytesOk = consumer.prefetchCount == 0 || consumer.activeCount < consumer.prefetchCount
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
	go consumer.consume(0)
}

func (consumer *Consumer) ping() {
	consumer.stopLock.Lock()
	defer consumer.stopLock.Unlock()
	if !consumer.stopped {
		select {
		case consumer.incoming <- true:
		default:
		}
	}

}

func (consumer *Consumer) consume(id uint16) {
	fmt.Printf("[C:%s#%d]Starting consumer\n", consumer.consumerTag, id)
	consumer.queue.maybeReady <- false
	for _ = range consumer.incoming {
		consumer.consumeOne()
	}
}

func (consumer *Consumer) consumeOne() {
	// Check local limit
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	if !consumer.consumerReady() {
		return
	}
	// Try to get message/check channel limit
	var msg = consumer.queue.getOne(consumer.channel, consumer)
	if msg == nil {
		return
	}
	var tag uint64 = 0
	if !consumer.noAck {
		tag = consumer.channel.addUnackedMessage(consumer, msg)
		consumer.incrActive(1, msg.size())
	}
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: msg.redelivered > 0,
		Exchange:    msg.exchange,
		RoutingKey:  msg.key,
	}, msg)
	consumer.statCount += 1
}

func (consumer *Consumer) consumeImmediate(msg *Message) bool {
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	if !consumer.consumerReady() {
		return false
	}
	var tag uint64 = 0
	if !consumer.noAck {
		tag = consumer.channel.addUnackedMessage(consumer, msg)
		consumer.incrActive(1, msg.size())
	}
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: msg.redelivered > 0,
		Exchange:    msg.exchange,
		RoutingKey:  msg.key,
	}, msg)
	consumer.statCount += 1
	return true
}

// Send again, leave all stats the same since this consumer was already
// dealing with this message
func (consumer *Consumer) redeliver(tag uint64, msg *Message) {
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: msg.redelivered > 0,
		Exchange:    msg.exchange,
		RoutingKey:  msg.key,
	}, msg)
}
