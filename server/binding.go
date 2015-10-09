package main

import (
	"github.com/jeffjenkins/mq/amqp"
)

type Binding struct {
	queueName    string
	exchangeName string
	key          string
	arguments    amqp.Table
}

func (b *Binding) matchDirect(message *amqp.BasicPublish) bool {
	return message.Exchange == b.exchangeName && b.key == message.RoutingKey
}

func (b *Binding) matchFanout(message *amqp.BasicPublish) bool {
	return true
}

func (b *Binding) matchTopic(message *amqp.BasicPublish) bool {
	panic("topic exchange not implemented")
}
