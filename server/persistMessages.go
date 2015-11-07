package main

import (
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/jeffjenkins/mq/amqp"
)

type MessageStore struct {
	// TODO: locks
	index    map[int64]*amqp.IndexMessage
	messages map[int64]*amqp.Message
	db       *bolt.DB
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

func (ms *MessageStore) addMessage(msg *amqp.Message, queues []string) error {
	// TODO: this is for testing durability, pull out the real value from
	// the header field
	durable := isDurable(msg)
	im := &amqp.IndexMessage{
		Id:      msg.Id,
		Refs:    int32(len(queues)),
		Durable: durable,
	}
	if durable {
		err := ms.db.Update(func(tx *bolt.Tx) error {
			persistMessage(tx, msg)
			persistIndexMessage(tx, im)
			for _, q := range queues {
				persistQueueMessage(tx, q, msg.Id)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	ms.index[msg.Id] = im
	ms.messages[msg.Id] = msg
	return nil
}

func (ms *MessageStore) addTxMsg(msgs []*amqp.TxMessage) error {
	// - Figure out of any messages are durable
	// - Create IndexMessage instances for each message id
	anyDurable := false
	indexMessages := make(map[int64]*amqp.IndexMessage)
	queuesNamesByMsg := make(map[int64][]string)
	for _, msg := range msgs {
		// calc any durable
		anyDurable = anyDurable || isDurable(msg.Msg)

		// calc index messages
		im, found := indexMessages[msg.Msg.Id]
		if !found {
			im := &amqp.IndexMessage{
				Id:      msg.Msg.Id,
				Refs:    0,
				Durable: isDurable(msg.Msg),
			}
			indexMessages[msg.Msg.Id] = im

		}
		im.Refs += 1

		// calc queues
		queues, found := queuesNamesByMsg[msg.Msg.Id]
		if !found {
			queues = make([]string, 0, 1)
		}
		queuesNamesByMsg[msg.Msg.Id] = append(queues, msg.QueueName)
	}

	// if any are durable, persist those ones
	if anyDurable {
		err := ms.db.Update(func(tx *bolt.Tx) error {
			for _, msg := range msgs {
				persistMessage(tx, msg.Msg)
				persistIndexMessage(tx, indexMessages[msg.Msg.Id])
				for _, q := range queuesNamesByMsg[msg.Msg.Id] {
					persistQueueMessage(tx, q, msg.Msg.Id)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	// Add to memory message store
	for _, msg := range msgs {
		ms.index[msg.Msg.Id] = indexMessages[msg.Msg.Id]
		ms.messages[msg.Msg.Id] = msg.Msg
	}
	return nil
}

func (ms *MessageStore) removeRef(msgId int64, queueName string) error {
	im, found := ms.index[msgId]
	if !found {
		panic("Integrity error: message in queue not in index")
	}
	// Update disk
	if im.Durable {
		err := ms.db.Update(func(tx *bolt.Tx) error {
			bId := make([]byte, 8)
			binary.PutVarint(bId, im.Id)
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
	im.Refs += 1
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

func persistMessage(tx *bolt.Tx, msg *amqp.Message) error {
	content_bucket, err := tx.CreateBucketIfNotExists([]byte("message_contents"))
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	bId := make([]byte, 8)
	binary.PutVarint(bId, msg.Id)
	return content_bucket.Put(bId, b)
}

func persistIndexMessage(tx *bolt.Tx, im *amqp.IndexMessage) error {
	content_bucket, err := tx.CreateBucketIfNotExists([]byte("message_index"))
	b, err := proto.Marshal(im)
	if err != nil {
		return err
	}
	bId := make([]byte, 8)
	binary.PutVarint(bId, im.Id)
	return content_bucket.Put(bId, b)
}

func persistQueueMessage(tx *bolt.Tx, queueName string, msgId int64) error {
	bucketName := fmt.Sprintf("queue_%s", queueName)
	content_bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	bId := make([]byte, 8)
	binary.PutVarint(bId, msgId)
	return content_bucket.Put(bId, make([]byte, 0))
}
