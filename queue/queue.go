package queue

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/consumer"
	"github.com/jeffjenkins/mq/interfaces"
	"github.com/jeffjenkins/mq/msgstore"
	"github.com/jeffjenkins/mq/stats"
	"sync"
	"time"
)

type Queue struct {
	Name            string
	Durable         bool
	exclusive       bool
	autoDelete      bool
	arguments       *amqp.Table
	Closed          bool
	objLock         sync.RWMutex
	queue           *list.List // int64
	queueLock       sync.Mutex
	consumerLock    sync.RWMutex
	consumers       []*consumer.Consumer // *Consumer
	currentConsumer int
	statCount       uint64
	maybeReady      chan bool
	soleConsumer    *consumer.Consumer
	ConnId          int64
	deleteActive    time.Time
	hasHadConsumers bool
	msgStore        *msgstore.MessageStore
	statProcOne     stats.Histogram
	deleteChan      chan *Queue
}

func NewQueue(
	name string,
	durable bool,
	exclusive bool,
	autoDelete bool,
	arguments *amqp.Table,
	connId int64,
	msgStore *msgstore.MessageStore,
	deleteChan chan *Queue,
) *Queue {
	return &Queue{
		Name:       name,
		Durable:    durable,
		exclusive:  exclusive,
		autoDelete: autoDelete,
		arguments:  arguments,
		ConnId:     connId,
		msgStore:   msgStore,
		deleteChan: deleteChan,
		// Fields that aren't passed in
		statProcOne: stats.MakeHistogram("queue-proc-one"),
		queue:       list.New(),
		consumers:   make([]*consumer.Consumer, 0, 1),
		maybeReady:  make(chan bool, 1),
	}
}

func equivalentQueues(q1 *Queue, q2 *Queue) bool {
	// Note: autodelete is not included since the spec says to ignore
	// the field if the queue is already created
	if q1.Name != q2.Name {
		return false
	}
	if q1.Durable != q2.Durable {
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

func (q *Queue) Len() uint32 {
	var l = q.queue.Len()
	if l < 0 {
		panic("Queue length overflow!")
	}
	return uint32(l)
}

func (q *Queue) ActiveConsumerCount() uint32 {
	// TODO(MUST): don't count consumers in the Channel.Flow state once
	// that is implemented
	return uint32(len(q.consumers))
}

func (q *Queue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":       q.Name,
		"durable":    q.Durable,
		"exclusive":  q.exclusive,
		"connId":     q.ConnId,
		"autoDelete": q.autoDelete,
		"size":       q.queue.Len(),
		"consumers":  q.consumers,
	})
}

func (q *Queue) LoadFromMsgStore(msgStore *msgstore.MessageStore) {
	queueList, err := msgStore.LoadQueueFromDisk(q.Name)
	if err != nil {
		panic("Integrity error reading queue from disk! " + err.Error())
	}
	q.queue = queueList
}

func (q *Queue) Close() {
	// This discards any messages which would be added. It does not
	// do cleanup
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	q.Closed = true
}

func (q *Queue) Purge() uint32 {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	return q.purgeNotThreadSafe()
}

func (q *Queue) purgeNotThreadSafe() uint32 {
	var length = q.queue.Len()
	q.queue.Init()
	return uint32(length)
}

func (q *Queue) Add(qm *amqp.QueueMessage) bool {
	// NOTE: I tried using consumeImmediate before adding things to the queue,
	// but it caused a pretty significant slowdown.
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	if !q.Closed {
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

func (q *Queue) ConsumeImmediate(qm *amqp.QueueMessage) bool {
	// TODO: randomize or round-robin through consumers
	q.consumerLock.RLock()
	defer q.consumerLock.RUnlock()
	for _, consumer := range q.consumers {
		var msg, acquired = q.msgStore.Get(qm, consumer.MessageResourceHolders())
		if acquired {
			consumer.ConsumeImmediate(qm, msg)
			return true
		}
	}
	return false
}

func (q *Queue) Delete(ifUnused bool, ifEmpty bool) (uint32, error) {
	// Lock
	if !q.Closed {
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

func (q *Queue) Readd(queueName string, msg *amqp.QueueMessage) {
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
	if q.soleConsumer != nil && q.soleConsumer.ConsumerTag == consumerTag {
		q.soleConsumer = nil
	}
	// remove from list
	for i, c := range q.consumers {
		if c.ConsumerTag == consumerTag {
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
		c.SendCancel()
		c.Stop()
	}
	q.consumers = make([]*consumer.Consumer, 0, 1)
}

func (q *Queue) AddConsumer(c *consumer.Consumer, exclusive bool) (uint16, error) {
	if q.Closed {
		return 0, nil
	}
	// Reset auto-delete
	q.deleteActive = time.Unix(0, 0)

	// Add consumer
	q.consumerLock.Lock()
	if exclusive {
		if len(q.consumers) == 0 {
			q.soleConsumer = c
		} else {
			return 403, fmt.Errorf("Exclusive access denied, %d consumers active", len(q.consumers))
		}
	}
	q.consumers = append(q.consumers, c)
	q.hasHadConsumers = true
	q.consumerLock.Unlock()
	return 0, nil
}

func (q *Queue) Start() {
	fmt.Println("Queue started!")
	go func() {
		for _ = range q.maybeReady {
			if q.Closed {
				fmt.Printf("Queue closed!\n")
				break
			}
			q.processOne()
		}
	}()
}

func (q *Queue) MaybeReady() chan bool {
	return q.maybeReady
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
		c.Ping()
	}
}

func (q *Queue) GetOneForced() *amqp.QueueMessage {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	if q.queue.Len() == 0 {
		return nil
	}
	qMsg := q.queue.Remove(q.queue.Front()).(*amqp.QueueMessage)
	return qMsg
}

func (q *Queue) GetOne(channel interfaces.MessageResourceHolder, consumer interfaces.MessageResourceHolder) (*amqp.QueueMessage, *amqp.Message) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	// Empty check
	if q.queue.Len() == 0 || q.Closed {
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
