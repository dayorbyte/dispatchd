package main

import (
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/jeffjenkins/mq/amqp"
)

type MessageStore struct {
	index    map[int64]*amqp.IndexMessage
	messages map[int64]*amqp.Message
	db       *bolt.DB
}

func (ms *MessageStore) addMessage(msg *amqp.Message, queues []string) error {
	// TODO: this is for testing durability, pull out the real value from
	// the header field
	durable := true
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
