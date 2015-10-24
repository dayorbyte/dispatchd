package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
)

func depersistQueue(tx *bolt.Tx, queue *Queue) error {
	bucket, err := tx.CreateBucketIfNotExists([]byte("queues"))
	if err != nil {
		return fmt.Errorf("create bucket: %s", err)
	}
	return bucket.Delete([]byte(queue.name))
}

func depersistBinding(tx *bolt.Tx, binding *Binding) error {
	var method = &amqp.QueueBind{
		Exchange:   binding.exchangeName,
		Queue:      binding.queueName,
		RoutingKey: binding.key,
		Arguments:  binding.arguments,
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte("bindings"))
	if err != nil {
		return fmt.Errorf("create bucket: %s", err)
	}
	var buffer = bytes.NewBuffer(make([]byte, 0, 50)) // TODO: don't I know the size?
	method.Write(buffer)
	// trim off the first four bytes, they're the class/method, which we
	// already know
	var value = buffer.Bytes()[4:]
	// bindings aren't named, so we hash the bytes we were given. I wonder
	// if we could make make the bytes the key and use no value?
	hash := sha1.New()
	hash.Write(value)
	return bucket.Delete([]byte(hash.Sum(nil)))
}
