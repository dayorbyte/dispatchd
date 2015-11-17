package consumer

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
	cchannel      interfaces.ConsumerChannel
	ConsumerTag   string
	exclusive     bool
	incoming      chan bool
	noAck         bool
	noLocal       bool
	qos           uint16
	cqueue        interfaces.ConsumerQueue
	queueName     string
	consumeLock   sync.Mutex
	limitLock     sync.Mutex
	prefetchSize  uint32
	prefetchCount uint16
	activeSize    uint32
	activeCount   uint16
	stopLock      sync.Mutex
	stopped       bool
	StatCount     uint64
	localId       int64
	// stats
	statConsumeOneGetOne stats.Histogram
	statConsumeOne       stats.Histogram
	statConsumeOneAck    stats.Histogram
	statConsumeOneSend   stats.Histogram
}

func NewConsumer(
	msgStore *msgstore.MessageStore,
	arguments *amqp.Table,
	cchannel interfaces.ConsumerChannel,
	consumerTag string,
	exclusive bool,
	noAck bool,
	noLocal bool,
	qos uint16,
	cqueue interfaces.ConsumerQueue,
	queueName string,
	prefetchSize uint32,
	prefetchCount uint16,
	localId int64,
) *Consumer {
	return &Consumer{
		msgStore:      msgStore,
		arguments:     arguments,
		cchannel:      cchannel,
		ConsumerTag:   consumerTag,
		exclusive:     exclusive,
		incoming:      make(chan bool, 1),
		noAck:         noAck,
		noLocal:       noLocal,
		qos:           qos,
		cqueue:        cqueue,
		queueName:     queueName,
		prefetchSize:  prefetchSize,
		prefetchCount: prefetchCount,
		localId:       localId,
		// stats
		statConsumeOneGetOne: stats.MakeHistogram("Consume-One-Get-One"),
		statConsumeOne:       stats.MakeHistogram("Consume-One-"),
		statConsumeOneAck:    stats.MakeHistogram("Consume-One-Ack"),
		statConsumeOneSend:   stats.MakeHistogram("Consume-One-Send"),
	}
}

func (consumer *Consumer) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"tag": consumer.ConsumerTag,
		"stats": map[string]interface{}{
			"total":             consumer.StatCount,
			"qos":               consumer.qos,
			"active_size_bytes": consumer.activeSize,
			"active_count":      consumer.activeCount,
		},
		"ack": !consumer.noAck,
	})
}

// TODO: make this a field that we construct on init
func (consumer *Consumer) MessageResourceHolders() []interfaces.MessageResourceHolder {
	return []interfaces.MessageResourceHolder{consumer, consumer.cchannel}
}

func (consumer *Consumer) Stop() {
	if !consumer.stopped {
		consumer.stopLock.Lock()
		consumer.stopped = true
		close(consumer.incoming)
		consumer.stopLock.Unlock()
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
	if !consumer.cchannel.FlowActive() {
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

func (consumer *Consumer) Start() {
	go consumer.consume(0)
}

func (consumer *Consumer) Ping() {
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
	// TODO: what is this doing?
	consumer.cqueue.MaybeReady() <- false
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
	var qm, msg = consumer.cqueue.GetOne(consumer.cchannel, consumer)
	stats.RecordHisto(consumer.statConsumeOneGetOne, start)
	if qm == nil {
		return
	}
	var tag uint64 = 0
	start = stats.Start()
	if !consumer.noAck {
		tag = consumer.cchannel.AddUnackedMessage(consumer.ConsumerTag, qm, consumer.queueName)
	} else {
		// We aren't expecting an ack, so this is the last time the message
		// will be referenced.
		var rhs = []interfaces.MessageResourceHolder{consumer.cchannel, consumer}
		err = consumer.msgStore.RemoveRef(qm, consumer.queueName, rhs)
		if err != nil {
			panic("Error getting queue message")
		}
	}
	stats.RecordHisto(consumer.statConsumeOneAck, start)
	start = stats.Start()
	consumer.cchannel.SendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.ConsumerTag,
		DeliveryTag: tag,
		Redelivered: qm.DeliveryCount > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
	stats.RecordHisto(consumer.statConsumeOneSend, start)
	consumer.StatCount += 1
}

func (consumer *Consumer) SendCancel() {
	consumer.cchannel.SendMethod(&amqp.BasicCancel{consumer.ConsumerTag, true})
}

func (consumer *Consumer) ConsumeImmediate(qm *amqp.QueueMessage, msg *amqp.Message) bool {
	consumer.consumeLock.Lock()
	defer consumer.consumeLock.Unlock()
	var tag uint64 = 0
	if !consumer.noAck {
		tag = consumer.cchannel.AddUnackedMessage(consumer.ConsumerTag, qm, consumer.queueName)
	}
	consumer.cchannel.SendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.ConsumerTag,
		DeliveryTag: tag,
		Redelivered: msg.Redelivered > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
	consumer.StatCount += 1
	return true
}

// Send again, leave all stats the same since this consumer was already
// dealing with this message
func (consumer *Consumer) Redeliver(tag uint64, qm *amqp.QueueMessage) {
	msg, found := consumer.msgStore.GetNoChecks(qm.Id)
	if !found {
		panic("Integrity error, message not found in message store")
	}
	consumer.cchannel.SendContent(&amqp.BasicDeliver{
		ConsumerTag: consumer.ConsumerTag,
		DeliveryTag: tag,
		Redelivered: msg.Redelivered > 0,
		Exchange:    msg.Exchange,
		RoutingKey:  msg.Key,
	}, msg)
}
