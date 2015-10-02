package main

type Message struct {
	header *amqp.ContentHeaderFrame
	payload []amqp.WireFrame
}

type Queue struct {
	name string
	passive bool
	durable bool
	exclusive bool
	autoDelete bool
	noWait bool
	arguments Table
	queue chan Message
}

func (q *Queue) add(message Message) {
	q.queue <- message
}

type Binding struct {
	queueName string
	exchangeName string
	routingKey string
	noWait bool
	arguments Table
}

func (b *Binding) matches(method *amqp.BasicPublish) bool {
	return false
}