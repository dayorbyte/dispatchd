package msgstore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/jeffjenkins/mq/amqp"
	"sync"
)

type MessageStore struct {
	index     map[int64]*amqp.IndexMessage
	messages  map[int64]*amqp.Message
	db        *bolt.DB
	msgLock   sync.RWMutex
	indexLock sync.RWMutex
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

func NewMessageStore(fileName string) (*MessageStore, error) {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &MessageStore{
		index:    make(map[int64]*amqp.IndexMessage),
		messages: make(map[int64]*amqp.Message),
		db:       db,
	}, nil
}

func (ms *MessageStore) Get(id int64) (msg *amqp.Message, found bool) {
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
		msgs = append(msgs, &amqp.TxMessage{
			Msg:       msg,
			QueueName: q,
		})
	}
	return ms.AddTxMessages(msgs)
}

func (ms *MessageStore) AddTxMessages(msgs []*amqp.TxMessage) (map[string][]*amqp.QueueMessage, error) {
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
			im = &amqp.IndexMessage{
				Id:      msg.Msg.Id,
				Refs:    0,
				Durable: isDurable(msg.Msg),
			}
			indexMessages[msg.Msg.Id] = im
		}
		im.Refs += 1

		// calc queues
		queues, found := queueMessages[msg.QueueName]
		if !found {
			queues = make([]*amqp.QueueMessage, 0, 1)
		}
		qm := &amqp.QueueMessage{
			Id:            msg.Msg.Id,
			DeliveryCount: 0,
			Durable:       msgDurable,
			MsgSize:       messageSize(msg.Msg),
		}
		queueMessages[msg.QueueName] = append(queues, qm)
	}

	// if any are durable, persist those ones
	if anyDurable {
		err := ms.db.Update(func(tx *bolt.Tx) error {
			// Save messages to content/index stores
			for _, msg := range msgs {
				persistMessage(tx, msg.Msg)
				persistIndexMessage(tx, indexMessages[msg.Msg.Id])
			}
			// Add messages to queues
			for q, qms := range queueMessages {
				for _, qm := range qms {
					persistQueueMessage(tx, q, qm)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	// Add to memory message store
	ms.msgLock.Lock()
	defer ms.msgLock.Unlock()
	ms.indexLock.Lock()
	defer ms.indexLock.Unlock()
	for _, msg := range msgs {
		ms.index[msg.Msg.Id] = indexMessages[msg.Msg.Id]
		ms.messages[msg.Msg.Id] = msg.Msg
	}
	return queueMessages, nil
}

func (ms *MessageStore) IncrDeliveryCount(queueName string, qm *amqp.QueueMessage) (err error) {
	qm.DeliveryCount += 1
	if qm.Durable {
		err = ms.db.Update(func(tx *bolt.Tx) error {
			persistQueueMessage(tx, queueName, qm)
			return nil
		})
	}
	return
}

func (ms *MessageStore) GetAndDecrRef(msgId int64, queueName string) (*amqp.Message, error) {
	msg, found := ms.Get(msgId)
	if !found {
		panic("Integrity error!")
	}
	if err := ms.RemoveRef(msgId, queueName); err != nil {
		return nil, err
	}
	return msg, nil
}

func (ms *MessageStore) RemoveRef(msgId int64, queueName string) error {
	im, found := ms.GetIndex(msgId)
	if !found {
		panic("Integrity error: message in queue not in index")
	}
	// Update disk
	if im.Durable {
		err := ms.db.Update(func(tx *bolt.Tx) error {
			bId := binaryId(im.Id)
			depersistQueueMessage(tx, queueName, bId)
			remaining, err := decrIndexMessage(tx, bId)
			if err != nil {
				return err
			}
			if remaining == 0 {
				return depersistMessage(tx, bId)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	// Update memory
	im.Refs -= 1
	if im.Refs == 0 {
		ms.msgLock.Lock()
		defer ms.msgLock.Unlock()
		ms.indexLock.Lock()
		defer ms.indexLock.Unlock()
		delete(ms.index, msgId)
		delete(ms.messages, msgId)
	}
	return nil
}

func depersistMessage(tx *bolt.Tx, bId []byte) error {
	content_bucket, err := tx.CreateBucketIfNotExists([]byte("message_contents"))
	if err != nil {
		return err
	}
	return content_bucket.Delete(bId)
}

func decrIndexMessage(tx *bolt.Tx, bId []byte) (int32, error) {
	// bucket
	content_bucket, err := tx.CreateBucketIfNotExists([]byte("message_index"))
	if err != nil {
		return -1, err
	}
	// get
	protoBytes := content_bucket.Get(bId)
	im := &amqp.IndexMessage{}
	err = proto.Unmarshal(protoBytes, im)
	if err != nil {
		return -1, err
	}
	// decr then save or delete
	if im.Refs <= 1 {
		content_bucket.Delete(bId)
		return 0, nil
	}
	im.Refs -= 1
	// TODO: panic on <0
	if im.Refs < 0 {
		return 0, content_bucket.Delete(bId)
	}
	newBytes, err := proto.Marshal(im)
	if err != nil {
		return -1, err
	}
	return im.Refs, content_bucket.Put(bId, newBytes)

}

func depersistQueueMessage(tx *bolt.Tx, queueName string, bId []byte) error {
	bucketName := fmt.Sprintf("queue_%s", queueName)
	content_bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	return content_bucket.Delete(bId)
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

func persistMessage(tx *bolt.Tx, msg *amqp.Message) error {
	content_bucket, err := tx.CreateBucketIfNotExists([]byte("message_contents"))
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return content_bucket.Put(binaryId(msg.Id), b)
}

func persistIndexMessage(tx *bolt.Tx, im *amqp.IndexMessage) error {
	content_bucket, err := tx.CreateBucketIfNotExists([]byte("message_index"))
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
