package interfaces

import (
	"github.com/jeffjenkins/mq/amqp"
)

// A message resource is something which has limits on the count of
// messages it can handle as well as the cumulative size of the messages
// it can handle.
type MessageResourceHolder interface {
	AcquireResources(qm *amqp.QueueMessage) bool
	ReleaseResources(qm *amqp.QueueMessage)
}

// The methods necessary for a consumer to interact with a channel
type ConsumerChannel interface {
	MessageResourceHolder
	SendContent(method amqp.MethodFrame, msg *amqp.Message)
	SendMethod(method amqp.MethodFrame)
	FlowActive() bool
	AddUnackedMessage(consumerTag string, qm *amqp.QueueMessage, queueName string) uint64
}

type ConsumerQueue interface {
	GetOne(
		channel MessageResourceHolder,
		consumer MessageResourceHolder,
	) (*amqp.QueueMessage, *amqp.Message)
	MaybeReady() chan bool
}
