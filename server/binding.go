package main

import (
	"github.com/jeffjenkins/mq/amqp"
	"regexp"
	"strings"
)

type Binding struct {
	queueName    string
	exchangeName string
	key          string
	arguments    amqp.Table
	topicMatcher *regexp.Regexp
}

func NewBinding(queueName string, exchangeName string, key string, arguments amqp.Table) *Binding {
	var parts = strings.Split(key, ".")
	for i, part := range parts {
		if part == "*" {
			parts[i] = `[^\.]+`
			continue
		}
		if part == "#" {
			parts[i] = ".*"
		}
	}
	// TODO: deal with failed compile
	expression := "^" + strings.Join(parts, `\.`) + "$"
	var regexp, success = regexp.Compile(expression)
	if success != nil {
		panic("Could not compile regex: '" + expression + "'")
	}
	return &Binding{
		queueName:    queueName,
		exchangeName: exchangeName,
		key:          key,
		arguments:    arguments,
		topicMatcher: regexp,
	}
}

func (b *Binding) matchDirect(message *amqp.BasicPublish) bool {
	return message.Exchange == b.exchangeName && b.key == message.RoutingKey
}

func (b *Binding) matchFanout(message *amqp.BasicPublish) bool {
	return true
}

func (b *Binding) matchTopic(message *amqp.BasicPublish) bool {
	var ex = b.exchangeName == message.Exchange
	var match = b.topicMatcher.MatchString(message.RoutingKey)
	return ex && match
}
