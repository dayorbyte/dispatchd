package msgstore

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/jeffjenkins/dispatchd/amqp"
	"github.com/jeffjenkins/dispatchd/persist"
	"github.com/jeffjenkins/dispatchd/stats"
	"sync"
	"time"
)

var MESSAGE_INDEX_BUCKET = []byte("message_index")
var MESSAGE_CONTENT_BUCKET = []byte("message_content")

type IndexMessageFactory struct{}

func (imf *IndexMessageFactory) New() proto.Unmarshaler {
	return &amqp.IndexMessage{}
}

type MessageContentFactory struct{}

func (mcf *MessageContentFactory) New() proto.Unmarshaler {
	return &amqp.Message{}
}

type QueueMessageFactory struct{}

func (qmf *QueueMessageFactory) New() proto.Unmarshaler {
	return &amqp.QueueMessage{}
}

const (
	opAdd           = iota
	opDelete        = iota
	opIncrDelivered = iota
)

type PersistKey struct {
	id        int64
	queueName string
}

type MessageStore struct {
	index         map[int64]*amqp.IndexMessage
	messages      map[int64]*amqp.Message
	addOps        map[PersistKey]*amqp.QueueMessage
	delOps        map[PersistKey]*amqp.QueueMessage
	deliveredOps  map[PersistKey]*amqp.QueueMessage
	persistLock   sync.Mutex
	db            *bolt.DB
	msgLock       sync.RWMutex
	indexLock     sync.RWMutex
	statAdd       stats.Histogram
	statRemoveRef stats.Histogram
}

func NewMessageStore(fileName string) (*MessageStore, error) {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		return nil, err
	}
	ms := &MessageStore{
		index:        make(map[int64]*amqp.IndexMessage),
		messages:     make(map[int64]*amqp.Message),
		db:           db,
		addOps:       make(map[PersistKey]*amqp.QueueMessage),
		delOps:       make(map[PersistKey]*amqp.QueueMessage),
		deliveredOps: make(map[PersistKey]*amqp.QueueMessage),
	}
	// Stats
	ms.statAdd = stats.MakeHistogram("add-message")
	ms.statRemoveRef = stats.MakeHistogram("remove-ref")

	return ms, nil
}

func (ms *MessageStore) Start() {
	go ms.periodicPersist()
}

func (ms *MessageStore) MessageCount() int {
	return len(ms.messages)
}

func (ms *MessageStore) IndexCount() int {
	return len(ms.index)
}

func messageSize(message *amqp.Message) uint32 {
	// TODO: include header size
	var size uint32 = 0
	for _, frame := range message.Payload {
		size += uint32(len(frame.Payload))
	}
	return size
}

func isDurable(msg *amqp.Message) bool {
	if msg == nil {
		panic("Message is nil(!!!)")
	}
	dm := msg.Header.Properties.DeliveryMode
	return dm != nil && *dm == byte(2)
}

func (ms *MessageStore) periodicPersist() {
	for {
		time.Sleep(200 * time.Millisecond)
		ms.persistOnce()
	}
}

func (ms *MessageStore) clearOps() {
	ms.delOps = make(map[PersistKey]*amqp.QueueMessage)
	ms.addOps = make(map[PersistKey]*amqp.QueueMessage)
	ms.deliveredOps = make(map[PersistKey]*amqp.QueueMessage)
}

func (ms *MessageStore) persistOnce() {
	// fmt.Println("Starting persist")
	// Snapshot so we can keep queueing persist ops
	ms.persistLock.Lock()
	delOps := ms.delOps
	addOps := ms.addOps
	deliveredOps := ms.deliveredOps
	ms.clearOps()
	ms.persistLock.Unlock()

	// We don't need to add or mark delivered anything we are going to delete
	// We don't need to delete anything we haven't added yet
	noDelete := make([]PersistKey, 0, len(addOps))
	for id, _ := range delOps {
		if _, ok := addOps[id]; ok {
			delete(addOps, id)
			noDelete = append(noDelete, id)
		}
		delete(deliveredOps, id)
	}
	for _, id := range noDelete {
		delete(delOps, id)
	}
	// fmt.Printf("Persist: add:%d, del:%d, delivery:%d\n", len(addOps), len(delOps), len(deliveredOps))
	err := ms.db.Update(func(tx *bolt.Tx) error {
		// Add
		msgsAdded := make(map[int64]bool)
		for pk, qm := range addOps {
			// Add -- Save messages to content/index stores
			if _, ok := msgsAdded[pk.id]; !ok {
				msg, okM := ms.GetNoChecks(pk.id)
				im, okI := ms.GetIndex(pk.id)
				if okM != okI {
					panic("Message index integrity error")
				}
				if !okM {
					// this message was deleted after we did a snapshot. We'll
					// still call a useless delete on it later, but we don't
					// need to add it now
					continue
				}
				persistMessage(tx, msg)
				persistIndexMessage(tx, im)
			}
			// Add -- Add messages to queues
			persistQueueMessage(tx, pk.queueName, qm)
		}

		// Update Delivered
		for pk, qm := range deliveredOps {
			persistQueueMessage(tx, pk.queueName, qm)
		}

		// Delete
		// Delete -- Remove from queue
		for pk, qm := range delOps {
			var err = depersistQueueMessage(tx, pk.queueName, qm.Id)
			if err != nil {
				return err
			}
			remaining, err := decrIndexMessage(tx, qm.Id, ms)
			if err != nil {
				return err
			}
			// Delete -- Delete message all together if there are no references left
			if remaining == 0 {
				return depersistMessage(tx, qm.Id)
			}
		}
		return nil
	})
	// TODO: this should probably just print a critical log message rather than
	//       killing the server
	if err != nil {
		panic("Failed to persist: " + err.Error())
	}
}

func (ms *MessageStore) LoadMessages() error {
	// Index
	imMap, err := persist.LoadAll(ms.db, MESSAGE_INDEX_BUCKET, &IndexMessageFactory{})
	if err != nil {
		return err
	}
	for _, unmarshaler := range imMap {
		var im = unmarshaler.(*amqp.IndexMessage)
		ms.index[im.Id] = im
	}
	// Content
	// TODO: don't load all content if it won't fit in memory
	mMap, err := persist.LoadAll(ms.db, MESSAGE_CONTENT_BUCKET, &MessageContentFactory{})
	if err != nil {
		return err
	}
	for _, unmarshaler := range mMap {
		var msg = unmarshaler.(*amqp.Message)
		ms.messages[msg.Id] = msg
	}
	return nil
}

func (ms *MessageStore) LoadQueueFromDisk(queueName string) (*list.List, error) { // list[amqp.QueueMessage]
	var ret = list.New()
	qmMap, err := persist.LoadAll(ms.db, []byte(fmt.Sprintf("queue_%s", queueName)), &QueueMessageFactory{})
	if err != nil {
		return nil, err
	}
	for _, unmarshaler := range qmMap {
		var qm = unmarshaler.(*amqp.QueueMessage)
		qm.LocalId = -1
		ret.PushFront(qm)
	}
	return ret, nil
}

func (ms *MessageStore) Fsck() ([]int64, []int64) {
	// TODO: make a function to find dangling or missing messages
	return make([]int64, 0), make([]int64, 0)
}

func (ms *MessageStore) Get(qm *amqp.QueueMessage, rhs []amqp.MessageResourceHolder) (*amqp.Message, bool) {
	ms.msgLock.RLock()
	defer ms.msgLock.RUnlock()
	// Acquire resources
	var acquired = make([]amqp.MessageResourceHolder, 0, len(rhs))
	for _, rh := range rhs {
		if !rh.AcquireResources(qm) {
			break
		}
		acquired = append(acquired, rh)
	}

	// Success! Return the message
	if len(acquired) == len(rhs) {
		var msg, found = ms.messages[qm.Id]
		if !found {
			panic("Integrity error! Message not found")
		}
		return msg, true
	}

	// Failure! Release the resources we already acquired
	for _, rh := range acquired {
		rh.ReleaseResources(qm)
	}
	return nil, false
}

func (ms *MessageStore) GetNoChecks(id int64) (msg *amqp.Message, found bool) {
	ms.msgLock.RLock()
	defer ms.msgLock.RUnlock()
	msg, found = ms.messages[id]
	return
}

func (ms *MessageStore) GetIndex(id int64) (msg *amqp.IndexMessage, found bool) {
	ms.indexLock.RLock()
	defer ms.indexLock.RUnlock()
	msg, found = ms.index[id]
	return

}

func (ms *MessageStore) AddMessage(msg *amqp.Message, queues []string) (map[string][]*amqp.QueueMessage, error) {
	msgs := make([]*amqp.TxMessage, 0, len(queues))
	for _, q := range queues {
		msgs = append(msgs, amqp.NewTxMessage(msg, q))
	}
	return ms.AddTxMessages(msgs)
}

func (ms *MessageStore) AddTxMessages(msgs []*amqp.TxMessage) (map[string][]*amqp.QueueMessage, error) {
	defer stats.RecordHisto(ms.statAdd, stats.Start())

	// - Figure out of any messages are durable
	// - Create IndexMessage instances for each message id
	anyDurable := false
	indexMessages := make(map[int64]*amqp.IndexMessage)
	queueMessages := make(map[string][]*amqp.QueueMessage)
	for _, msg := range msgs {
		// calc any durable
		var msgDurable = isDurable(msg.Msg)
		anyDurable = anyDurable || msgDurable
		// calc index messages
		var im, found = indexMessages[msg.Msg.Id]
		if !found {
			im = amqp.NewIndexMessage(msg.Msg.Id, 0, isDurable(msg.Msg), 0)
			indexMessages[msg.Msg.Id] = im
		}
		im.Refs += 1

		// calc queues
		queues, found := queueMessages[msg.QueueName]
		if !found {
			queues = make([]*amqp.QueueMessage, 0, 1)
		}
		qm := amqp.NewQueueMessage(
			msg.Msg.Id,
			0,
			msgDurable,
			messageSize(msg.Msg),
			msg.Msg.LocalId,
		)
		queueMessages[msg.QueueName] = append(queues, qm)
	}
	// if any are durable, persist those ones
	if anyDurable {
		ms.persistLock.Lock()
		for q, qms := range queueMessages {
			for _, qm := range qms {
				ms.addOps[PersistKey{qm.Id, q}] = qm
			}
		}
		ms.persistLock.Unlock()
	}
	// Add to memory message store
	ms.msgLock.Lock()
	defer ms.msgLock.Unlock()
	ms.indexLock.Lock()
	defer ms.indexLock.Unlock()
	for _, msg := range msgs {
		// fmt.Printf("Adding to index: %d\n", msg.Msg.Id)
		ms.index[msg.Msg.Id] = indexMessages[msg.Msg.Id]
		ms.messages[msg.Msg.Id] = msg.Msg
	}
	return queueMessages, nil
}

func (ms *MessageStore) IncrDeliveryCount(queueName string, qm *amqp.QueueMessage) (err error) {
	qm.DeliveryCount += 1
	if qm.Durable {
		ms.persistLock.Lock()
		ms.deliveredOps[PersistKey{qm.Id, queueName}] = qm
		ms.persistLock.Unlock()
	}
	return
}

func (ms *MessageStore) GetAndDecrRef(qm *amqp.QueueMessage, queueName string, rhs []amqp.MessageResourceHolder) (*amqp.Message, error) {
	msg, found := ms.GetNoChecks(qm.Id)
	if !found {
		panic("Integrity error!")
	}
	if err := ms.RemoveRef(qm, queueName, rhs); err != nil {
		return nil, err
	}
	return msg, nil
}

func (ms *MessageStore) RemoveRef(qm *amqp.QueueMessage, queueName string, rhs []amqp.MessageResourceHolder) error {
	defer stats.RecordHisto(ms.statRemoveRef, stats.Start())
	im, found := ms.GetIndex(qm.Id)
	if !found {
		panic("Integrity error: message in queue not in index")
	}
	if len(queueName) == 0 {
		panic("Bad queue name!")
	}
	// Update disk
	if im.Durable {
		ms.persistLock.Lock()
		ms.delOps[PersistKey{im.Id, queueName}] = qm
		ms.persistLock.Unlock()
	} else {
		// Update if only memory
		im.Refs -= 1
		if im.Refs == 0 {
			ms.msgLock.Lock()
			delete(ms.index, qm.Id)
			ms.msgLock.Unlock()

			ms.indexLock.Lock()
			delete(ms.messages, qm.Id)
			ms.indexLock.Unlock()
		}
	}

	for _, rh := range rhs {
		rh.ReleaseResources(qm)
	}
	return nil
}

func depersistMessage(tx *bolt.Tx, id int64) error {
	content_bucket, err := tx.CreateBucketIfNotExists(MESSAGE_CONTENT_BUCKET)
	if err != nil {
		return err
	}
	return content_bucket.Delete(binaryId(id))
}

func decrIndexMessage(tx *bolt.Tx, id int64, ms *MessageStore) (int32, error) {
	// bucket
	index_bucket, err := tx.CreateBucketIfNotExists(MESSAGE_INDEX_BUCKET)
	if err != nil {
		return -1, err
	}
	var bId = binaryId(id)
	// get
	protoBytes := index_bucket.Get(bId)
	im := &amqp.IndexMessage{}
	err = proto.Unmarshal(protoBytes, im)
	if err != nil {
		return -1, err
	}
	// decr then save or delete
	if im.Refs < 1 {
		panic("Index message would have gone negative!")
		// TODO: isn't this a data integrity error?
		index_bucket.Delete(bId)
		return 0, nil
	}
	im.Refs -= 1
	// TODO: panic on <0
	if im.Refs == 0 {
		ms.msgLock.Lock()
		delete(ms.index, id)
		ms.msgLock.Unlock()

		ms.indexLock.Lock()
		delete(ms.messages, id)
		ms.indexLock.Unlock()
		return 0, index_bucket.Delete(bId)
	}
	newBytes, err := proto.Marshal(im)
	if err != nil {
		return -1, err
	}
	return im.Refs, index_bucket.Put(bId, newBytes)

}

func depersistQueueMessage(tx *bolt.Tx, queueName string, id int64) error {
	bucketName := fmt.Sprintf("queue_%s", queueName)
	content_bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	var key = binaryId(id)
	return content_bucket.Delete(key)
}

func binaryId(id int64) []byte {
	var buf = bytes.NewBuffer(make([]byte, 0, 8))
	binary.Write(buf, binary.LittleEndian, id)
	var ret = buf.Bytes()
	if len(ret) != 8 {
		panic("Bad bytes!")
	}
	return ret
}

func bytesToInt64(bId []byte) int64 {
	var id int64
	buf := bytes.NewBuffer(bId)
	binary.Read(buf, binary.LittleEndian, &id)
	return id
}

func persistMessage(tx *bolt.Tx, msg *amqp.Message) error {
	content_bucket, err := tx.CreateBucketIfNotExists(MESSAGE_CONTENT_BUCKET)
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return content_bucket.Put(binaryId(msg.Id), b)
}

func persistIndexMessage(tx *bolt.Tx, im *amqp.IndexMessage) error {
	content_bucket, err := tx.CreateBucketIfNotExists(MESSAGE_INDEX_BUCKET)
	b, err := proto.Marshal(im)
	if err != nil {
		return err
	}
	return content_bucket.Put(binaryId(im.Id), b)
}

func persistQueueMessage(tx *bolt.Tx, queueName string, qm *amqp.QueueMessage) error {
	bucketName := fmt.Sprintf("queue_%s", queueName)
	content_bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	protoBytes, err := proto.Marshal(qm)
	if err != nil {
		return err
	}
	return content_bucket.Put(binaryId(qm.Id), protoBytes)
}
