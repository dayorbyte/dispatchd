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
