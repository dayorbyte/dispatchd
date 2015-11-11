package main

import (
	"encoding/json"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
)

type Consumer struct {
	arguments     *amqp.Table
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
	consumer.queue.maybeReady <- false
	for _ = range consumer.incoming {
		consumer.consumeOne()
	}
}

func (consumer *Consumer) consumeOne() {
	var err error
	// Check local limit
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	if !consumer.consumerReady() {
		return
	}
	// Try to get message/check channel limit
	var qm = consumer.queue.getOne(consumer.channel, consumer)
	if qm == nil {
		return
	}
	var tag uint64 = 0
	var msg *amqp.Message
	var found bool
	if !consumer.noAck {
		tag = consumer.channel.addUnackedMessage(consumer, qm)
		// We get the message out without decrementing the ref because we're
		// expecting an ack. The ack code will decrement.
		msg, found = consumer.channel.server.msgStore.Get(qm.Id)
		if !found {
			panic("Integrity error, message id not found")
		}
		consumer.incrActive(1, qm.MsgSize)
	} else {
		// We aren't expecting an ack, so this is the last time the message
		// will be referenced.
		msg, err = consumer.channel.server.msgStore.GetAndDecrRef(qm.Id, consumer.queue.name)
		if err != nil {
			panic("Error getting queue message")
		}
	}
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: qm.DeliveryCount > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
	consumer.statCount += 1
}

func (consumer *Consumer) consumeImmediate(qm *amqp.QueueMessage) bool {
	var err error
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	if !consumer.consumerReady() {
		return false
	}
	var tag uint64 = 0
	var msg *amqp.Message
	if !consumer.noAck {
		tag = consumer.channel.addUnackedMessage(consumer, qm)
		consumer.incrActive(1, qm.MsgSize)
	} else {
		msg, err = consumer.channel.server.msgStore.GetAndDecrRef(qm.Id, consumer.queue.name)
		if err != nil {
			panic("Error getting queue message")
		}
	}
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: msg.Redelivered > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
	consumer.statCount += 1
	return true
}

// Send again, leave all stats the same since this consumer was already
// dealing with this message
func (consumer *Consumer) redeliver(tag uint64, qm *amqp.QueueMessage) {
	msg, found := consumer.channel.server.msgStore.Get(qm.Id)
	if !found {
		panic("Integrity error, message not found in message store")
	}
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: msg.Redelivered > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
}
