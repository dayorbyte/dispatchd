package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/interfaces"
	"github.com/jeffjenkins/mq/msgstore"
	"github.com/jeffjenkins/mq/stats"
	"sync"
	"time"
)

func NewMessage(method *amqp.BasicPublish, localId int64) *amqp.Message {
	return &amqp.Message{
		Id:       nextId(),
		Method:   method,
		Exchange: method.Exchange,
		Key:      method.RoutingKey,
		Payload:  make([]*amqp.WireFrame, 0, 1),
		LocalId:  localId,
	}
}

type Queue struct {
	name            string
	durable         bool
	exclusive       bool
	autoDelete      bool
	arguments       *amqp.Table
	closed          bool
	objLock         sync.RWMutex
	queue           *list.List // int64
	queueLock       sync.Mutex
	consumerLock    sync.RWMutex
	consumers       []*Consumer // *Consumer
	currentConsumer int
	statCount       uint64
	maybeReady      chan bool
	soleConsumer    *Consumer
	connId          int64
	deleteActive    time.Time
	hasHadConsumers bool
	msgStore        *msgstore.MessageStore
	statProcOne     stats.Histogram
	deleteChan      chan *Queue
}

func equivalentQueues(q1 *Queue, q2 *Queue) bool {
	// Note: autodelete is not included since the spec says to ignore
	// the field if the queue is already created
	if q1.name != q2.name {
		return false
	}
	if q1.durable != q2.durable {
		return false
	}
	if q1.exclusive != q2.exclusive {
		return false
	}
	if !amqp.EquivalentTables(q1.arguments, q2.arguments) {
		return false
	}
	return true
}

func (q *Queue) activeConsumerCount() uint32 {
	// TODO(MUST): don't count consumers in the Channel.Flow state once
	// that is implemented
	return uint32(len(q.consumers))
}

func (q *Queue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":       q.name,
		"durable":    q.durable,
		"exclusive":  q.exclusive,
		"connId":     q.connId,
		"autoDelete": q.autoDelete,
		"size":       q.queue.Len(),
		"consumers":  q.consumers,
	})
}

func (q *Queue) close() {
	// This discards any messages which would be added. It does not
	// do cleanup
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	q.closed = true
}

func (q *Queue) purge() uint32 {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	return q.purgeNotThreadSafe()
}

func (q *Queue) purgeNotThreadSafe() uint32 {
	var length = q.queue.Len()
	q.queue.Init()
	return uint32(length)
}

func (q *Queue) add(qm *amqp.QueueMessage) bool {
	// NOTE: I tried using consumeImmediate before adding things to the queue,
	// but it caused a pretty significant slowdown.
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	if !q.closed {
		q.statCount += 1
		q.queue.PushBack(qm)
		select {
		case q.maybeReady <- true:
		default:
		}
		return true
	} else {
		return false
	}
}

func (q *Queue) consumeImmediate(qm *amqp.QueueMessage) bool {
	// TODO: randomize or round-robin through consumers
	q.consumerLock.RLock()
	defer q.consumerLock.RUnlock()
	for _, consumer := range q.consumers {
		var rhs = []interfaces.MessageResourceHolder{
			consumer.channel,
			consumer,
		}
		var msg, acquired = q.msgStore.Get(qm, rhs)
		if acquired {
			consumer.consumeImmediate(qm, msg)
			return true
		}
	}
	return false
}

func (q *Queue) delete(ifUnused bool, ifEmpty bool) (uint32, error) {
	// Lock
	if !q.closed {
		panic("Queue deleted before it was closed!")
	}
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	// Check
	var usedOk = !ifUnused || len(q.consumers) == 0
	var emptyOk = !ifEmpty || q.queue.Len() == 0
	if !usedOk {
		return 0, errors.New("if-unused specified and there are consumers")
	}
	if !emptyOk {
		return 0, errors.New("if-empty specified and there are messages in the queue")
	}
	// Purge
	q.cancelConsumers()
	return q.purgeNotThreadSafe(), nil
}

func (q *Queue) readd(queueName string, msg *amqp.QueueMessage) {
	// TODO: if there is a consumer available, dispatch
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	// this method is only called when we get a nack or we shut down a channel,
	// so it means the message was not acked.
	q.msgStore.IncrDeliveryCount(queueName, msg)
	q.queue.PushFront(msg)
	q.maybeReady <- true
}

func (q *Queue) removeConsumer(consumerTag string) {
	q.consumerLock.Lock()
	defer q.consumerLock.Unlock()
	if q.soleConsumer != nil && q.soleConsumer.consumerTag == consumerTag {
		q.soleConsumer = nil
	}
	// remove from list
	for i, c := range q.consumers {
		if c.consumerTag == consumerTag {
			q.consumers = append(q.consumers[:i], q.consumers[i+1:]...)
		}
	}
	var size = len(q.consumers)
	if size == 0 {
		q.currentConsumer = 0
		if q.autoDelete && q.hasHadConsumers {
			go q.autodeleteTimeout()
		}
	} else {
		q.currentConsumer = q.currentConsumer % size
	}

}

func (q *Queue) autodeleteTimeout() {
	// There's technically a race condition here where a new binding could be
	// added right as we check this, but after a 5 second wait with no activity
	// I think this is probably safe enough.
	var now = time.Now()
	q.deleteActive = now
	time.Sleep(5 * time.Second)
	if q.deleteActive == now {
		q.deleteChan <- q
	}
}

func (q *Queue) cancelConsumers() {
	q.consumerLock.Lock()
	defer q.consumerLock.Unlock()
	q.soleConsumer = nil
	// Send cancel to each consumer
	for _, c := range q.consumers {
		c.channel.sendMethod(&amqp.BasicCancel{c.consumerTag, true})
		c.stop()
	}
	q.consumers = make([]*Consumer, 0, 1)
}

func (q *Queue) addConsumer(channel *Channel, method *amqp.BasicConsume) (uint16, error) {
	if q.closed {
		return 0, nil
	}
	// Reset auto-delete
	q.deleteActive = time.Unix(0, 0)

	// Add consumer
	var consumer = &Consumer{
		msgStore:    q.msgStore,
		arguments:   method.Arguments,
		channel:     channel,
		consumerTag: method.ConsumerTag,
		exclusive:   method.Exclusive,
		incoming:    make(chan bool, 1),
		noAck:       method.NoAck,
		noLocal:     method.NoLocal,
		// TODO: size QOS
		qos:           channel.defaultPrefetchCount,
		queue:         q,
		prefetchSize:  channel.defaultPrefetchSize,
		prefetchCount: channel.defaultPrefetchCount,
		localId:       channel.conn.id,
		// stats
		statConsumeOneGetOne: stats.MakeHistogram("Consume-One-Get-One"),
		statConsumeOne:       stats.MakeHistogram("Consume-One-"),
		statConsumeOneAck:    stats.MakeHistogram("Consume-One-Ack"),
		statConsumeOneSend:   stats.MakeHistogram("Consume-One-Send"),
	}
	q.consumerLock.Lock()
	if method.Exclusive {
		if len(q.consumers) == 0 {
			q.soleConsumer = consumer
		} else {
			return 403, fmt.Errorf("Exclusive access denied, %d consumers active", len(q.consumers))
		}
	}
	channel.addConsumer(consumer)
	q.consumers = append(q.consumers, consumer)
	q.hasHadConsumers = true
	q.consumerLock.Unlock()
	consumer.start()
	return 0, nil
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
	defer stats.RecordHisto(q.statProcOne, stats.Start())
	q.consumerLock.RLock()
	defer q.consumerLock.RUnlock()
	var size = len(q.consumers)
	if size == 0 {
		return
	}
	for count := 0; count < size; count++ {
		q.currentConsumer = (q.currentConsumer + 1) % size
		var c = q.consumers[q.currentConsumer]
		c.ping()
	}
}

func (q *Queue) getOneForced() *amqp.QueueMessage {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	if q.queue.Len() == 0 {
		return nil
	}
	qMsg := q.queue.Remove(q.queue.Front()).(*amqp.QueueMessage)
	return qMsg
}

func (q *Queue) getOne(channel *Channel, consumer *Consumer) (*amqp.QueueMessage, *amqp.Message) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	// Empty check
	if q.queue.Len() == 0 || q.closed {
		return nil, nil
	}

	// Get one message. If there is a message try to acquire the resources
	// from the channel.
	var qm = q.queue.Front().Value.(*amqp.QueueMessage)

	var rhs = []interfaces.MessageResourceHolder{
		channel,
		consumer,
	}
	var msg, acquired = q.msgStore.Get(qm, rhs)
	if acquired {
		q.queue.Remove(q.queue.Front())
		return qm, msg
	}
	return nil, nil
}
