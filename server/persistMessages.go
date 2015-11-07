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
				persisQueueMessage(tx, q, msg.Id)
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

func persisQueueMessage(tx *bolt.Tx, queueName string, msgId int64) error {
	bucketName := fmt.Sprintf("queue_%s", queueName)
	content_bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	bId := make([]byte, 8)
	binary.PutVarint(bId, msgId)
	return content_bucket.Put(bId, make([]byte, 0))
}
