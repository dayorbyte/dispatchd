package main

import (
	"encoding/json"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/interfaces"
	"github.com/jeffjenkins/mq/msgstore"
	"github.com/jeffjenkins/mq/stats"
	"sync"
)

type Consumer struct {
	msgStore      *msgstore.MessageStore
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
	// stats
	statConsumeOneGetOne stats.Histogram
	statConsumeOne       stats.Histogram
	statConsumeOneAck    stats.Histogram
	statConsumeOneSend   stats.Histogram
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

func (consumer *Consumer) AcquireResources(qm *amqp.QueueMessage) bool {
	consumer.limitLock.Lock()
	defer consumer.limitLock.Unlock()

	// If no-local was set on the consumer, reject messages
	if consumer.noLocal && qm.LocalId == consumer.localId {
		return false
	}

	// If the channel is in flow mode we don't consume
	// TODO: If flow is mostly for producers, then maybe we
	// should consume? I feel like the right answer here is for
	// clients to not produce and consume on the same channel.
	if !consumer.channel.flow {
		return false
	}
	// If we aren't acking then there are no resource limits. Up the stats
	// and return true
	if consumer.noAck {
		consumer.activeCount += 1
		consumer.activeSize += qm.MsgSize
		return true
	}

	// Calculate whether we're over either of the size and count limits
	var sizeOk = consumer.prefetchSize == 0 || consumer.activeSize < consumer.prefetchSize
	var countOk = consumer.prefetchCount == 0 || consumer.activeCount < consumer.prefetchCount
	if sizeOk && countOk {
		consumer.activeCount += 1
		consumer.activeSize += qm.MsgSize
		return true
	}
	return false
}

func (consumer *Consumer) ReleaseResources(qm *amqp.QueueMessage) {
	consumer.limitLock.Lock()
	consumer.activeCount -= 1
	consumer.activeSize -= qm.MsgSize
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
	defer stats.RecordHisto(consumer.statConsumeOne, stats.Start())
	var err error
	// Check local limit
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	// Try to get message/check channel limit

	var start = stats.Start()
	var qm, msg = consumer.queue.getOne(consumer.channel, consumer)
	stats.RecordHisto(consumer.statConsumeOneGetOne, start)
	if qm == nil {
		return
	}
	var tag uint64 = 0
	start = stats.Start()
	if !consumer.noAck {
		tag = consumer.channel.addUnackedMessage(consumer, qm, consumer.queue.name)
	} else {
		// We aren't expecting an ack, so this is the last time the message
		// will be referenced.
		var rhs = []interfaces.MessageResourceHolder{consumer.channel, consumer}
		err = consumer.msgStore.RemoveRef(qm, consumer.queue.name, rhs)
		if err != nil {
			panic("Error getting queue message")
		}
	}
	stats.RecordHisto(consumer.statConsumeOneAck, start)
	start = stats.Start()
	consumer.channel.sendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.consumerTag,
		DeliveryTag: tag,
		Redelivered: qm.DeliveryCount > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
	stats.RecordHisto(consumer.statConsumeOneSend, start)
	consumer.statCount += 1
}

func (consumer *Consumer) consumeImmediate(qm *amqp.QueueMessage, msg *amqp.Message) bool {
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	var tag uint64 = 0
	if !consumer.noAck {
		tag = consumer.channel.addUnackedMessage(consumer, qm, consumer.queue.name)
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
	msg, found := consumer.msgStore.GetNoChecks(qm.Id)
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
