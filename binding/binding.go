package binding

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
	"regexp"
	"strings"
)

type Binding struct {
	QueueName    string
	ExchangeName string
	Key          string
	Arguments    *amqp.Table
	topicMatcher *regexp.Regexp
}

var topicRoutingPatternPattern, _ = regexp.Compile(`^((\w+|\*|#)(\.(\w+|\*|#))*|)$`)

func (binding *Binding) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"queueName":    binding.QueueName,
		"exchangeName": binding.ExchangeName,
		"key":          binding.Key,
		"arguments":    binding.Arguments,
	})
}

func (binding *Binding) Equals(other *Binding) bool {
	if other == nil || binding == nil {
		return false
	}
	return binding.QueueName == other.QueueName &&
		binding.ExchangeName == other.ExchangeName &&
		binding.Key == other.Key
}

func (binding *Binding) Depersist(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		return binding.DepersistBoltTx(tx)
	})
}

func (binding *Binding) DepersistBoltTx(tx *bolt.Tx) error {
	var method = &amqp.QueueBind{
		Exchange:   binding.ExchangeName,
		Queue:      binding.QueueName,
		RoutingKey: binding.Key,
		Arguments:  binding.Arguments,
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

func NewBinding(queueName string, exchangeName string, key string, arguments *amqp.Table, topic bool) (*Binding, error) {
	var re *regexp.Regexp = nil
	if topic {
		if !topicRoutingPatternPattern.MatchString(key) {
			return nil, fmt.Errorf("Topic exchange routing key can only have a-zA-Z0-9, or # or *")
		}
		var parts = strings.Split(key, ".")
		for i, part := range parts {
			if part == "*" {
				parts[i] = `[^\.]+`
			} else if part == "#" {
				parts[i] = ".*"
			} else {
				parts[i] = regexp.QuoteMeta(parts[i])
			}
		}
		expression := "^" + strings.Join(parts, `\.`) + "$"
		var err error = nil
		re, err = regexp.Compile(expression)
		if err != nil {
			return nil, fmt.Errorf("Could not compile regex: '%s'", expression)
		}
	}

	return &Binding{
		QueueName:    queueName,
		ExchangeName: exchangeName,
		Key:          key,
		Arguments:    arguments,
		topicMatcher: re,
	}, nil
}

func (b *Binding) MatchDirect(message *amqp.BasicPublish) bool {
	return message.Exchange == b.ExchangeName && b.Key == message.RoutingKey
}

func (b *Binding) MatchFanout(message *amqp.BasicPublish) bool {
	return true
}

func (b *Binding) MatchTopic(message *amqp.BasicPublish) bool {
	var ex = b.ExchangeName == message.Exchange
	var match = b.topicMatcher.MatchString(message.RoutingKey)
	return ex && match
}
