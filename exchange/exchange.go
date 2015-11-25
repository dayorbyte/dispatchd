package exchange

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/binding"
	"github.com/jeffjenkins/mq/queue"
	"sync"
	"time"
)

type Extype uint8

const (
	EX_TYPE_DIRECT  Extype = 1
	EX_TYPE_FANOUT  Extype = 2
	EX_TYPE_TOPIC   Extype = 3
	EX_TYPE_HEADERS Extype = 4
)

type Exchange struct {
	Name         string
	Extype       Extype
	Durable      bool
	autodelete   bool
	internal     bool
	arguments    *amqp.Table
	System       bool
	bindings     []*binding.Binding
	bindingsLock sync.Mutex
	incoming     chan amqp.Frame
	Closed       bool
	deleteActive time.Time
	deleteChan   chan *Exchange
}

func (exchange *Exchange) Close() {
	exchange.Closed = true
}

func (exchange *Exchange) MarshalJSON() ([]byte, error) {
	var typ, err = exchangeTypeToName(exchange.Extype)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"type":     typ,
		"bindings": exchange.bindings,
	})
}

func NewExchange(
	name string,
	extype Extype,
	durable bool,
	autodelete bool,
	internal bool,
	arguments *amqp.Table,
	system bool,
	deleteChan chan *Exchange,
) *Exchange {
	return &Exchange{
		Name:       name,
		Extype:     extype,
		Durable:    durable,
		autodelete: autodelete,
		internal:   internal,
		arguments:  arguments,
		System:     system,
		deleteChan: deleteChan,
		// not passed in
		incoming: make(chan amqp.Frame),
		bindings: make([]*binding.Binding, 0),
	}
}

func (ex1 *Exchange) EquivalentExchanges(ex2 *Exchange) bool {
	// NOTE: auto-delete is ignored for existing exchanges, so we
	// do not check it here.
	if ex1.Name != ex2.Name {
		return false
	}
	if ex1.Extype != ex2.Extype {
		return false
	}
	if ex1.Durable != ex2.Durable {
		return false
	}
	if ex1.internal != ex2.internal {
		return false
	}
	if !amqp.EquivalentTables(ex1.arguments, ex2.arguments) {
		return false
	}
	return true
}

func ExchangeNameToType(et string) (Extype, error) {
	switch {
	case et == "direct":
		return EX_TYPE_DIRECT, nil
	case et == "fanout":
		return EX_TYPE_FANOUT, nil
	case et == "topic":
		return EX_TYPE_TOPIC, nil
	case et == "headers":
		return EX_TYPE_HEADERS, nil
	default:
		return 0, fmt.Errorf("Unknown exchang type '%s', %d %d", et, len(et), len("direct"))
	}
}

func exchangeTypeToName(et Extype) (string, error) {
	switch {
	case et == EX_TYPE_DIRECT:
		return "direct", nil
	case et == EX_TYPE_FANOUT:
		return "fanout", nil
	case et == EX_TYPE_TOPIC:
		return "topic", nil
	case et == EX_TYPE_HEADERS:
		return "headers", nil
	default:
		return "", fmt.Errorf("bad exchange type: %d", et)
	}
}

func (exchange *Exchange) QueuesForPublish(msg *amqp.Message) (map[string]bool, *amqp.AMQPError) {
	var queues = make(map[string]bool)
	switch {
	case exchange.Extype == EX_TYPE_DIRECT:
		for _, binding := range exchange.bindings {
			if binding.MatchDirect(msg.Method) {
				var _, alreadySeen = queues[binding.QueueName]
				if alreadySeen {
					continue
				}
				queues[binding.QueueName] = true
			}
		}
	case exchange.Extype == EX_TYPE_FANOUT:
		for _, binding := range exchange.bindings {
			if binding.MatchFanout(msg.Method) {
				var _, alreadySeen = queues[binding.QueueName]
				if alreadySeen {
					continue
				}
				queues[binding.QueueName] = true
			}
		}
	case exchange.Extype == EX_TYPE_TOPIC:
		for _, binding := range exchange.bindings {
			if binding.MatchTopic(msg.Method) {
				var _, alreadySeen = queues[binding.QueueName]
				if alreadySeen {
					continue
				}
				queues[binding.QueueName] = true
			}
		}
	// case exchange.Extype == EX_TYPE_HEADERS:
	// 	// TODO: implement
	// 	panic("Headers is not implemented!")
	default: // pragma: nocover
		panic("Unknown exchange type created somehow. Server integrity error!")
	}
	return queues, nil
}

func (exchange *Exchange) Depersist(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, binding := range exchange.bindings {
			if err := binding.DepersistBoltTx(tx); err != nil {
				return err
			}
		}
		return DepersistExchangeBoltTx(tx, exchange)
	})
}

func DepersistExchangeBoltTx(tx *bolt.Tx, exchange *Exchange) error {
	bucket, err := tx.CreateBucketIfNotExists([]byte("exchanges"))
	if err != nil {
		return fmt.Errorf("create bucket: %s", err)
	}
	return bucket.Delete([]byte(exchange.Name))
}

func (exchange *Exchange) IsTopic() bool {
	return exchange.Extype == EX_TYPE_TOPIC
}

func (exchange *Exchange) AddBinding(method *amqp.QueueBind, connId int64, fromDisk bool) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	var binding, err = binding.NewBinding(method.Queue, method.Exchange, method.RoutingKey, method.Arguments, exchange.IsTopic())
	if err != nil {
		return err
	}

	for _, b := range exchange.bindings {
		if binding.Equals(b) {
			return nil
		}
	}

	if exchange.autodelete {
		exchange.deleteActive = time.Unix(0, 0)
	}
	exchange.bindings = append(exchange.bindings, binding)
	return nil
}

func (exchange *Exchange) BindingsForQueue(queueName string) []*binding.Binding {
	var ret = make([]*binding.Binding, 0)
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	for _, b := range exchange.bindings {
		if b.QueueName == queueName {
			ret = append(ret, b)
		}
	}
	return ret
}

func (exchange *Exchange) RemoveBindingsForQueue(queueName string) {
	var remaining = make([]*binding.Binding, 0)
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()
	for _, b := range exchange.bindings {
		if b.QueueName != queueName {
			remaining = append(remaining, b)
		}
	}
	exchange.bindings = remaining
}

func (exchange *Exchange) RemoveBinding(queue *queue.Queue, binding *binding.Binding) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	// Delete binding
	for i, b := range exchange.bindings {
		if binding.Equals(b) {
			exchange.bindings = append(exchange.bindings[:i], exchange.bindings[i+1:]...)
			if exchange.autodelete && len(exchange.bindings) == 0 {
				go exchange.autodeleteTimeout()
			}
			return nil
		}
	}
	return nil
}

func (exchange *Exchange) autodeleteTimeout() {
	// There's technically a race condition here where a new binding could be
	// added right as we check this, but after a 5 second wait with no activity
	// I think this is probably safe enough.
	var now = time.Now()
	exchange.deleteActive = now
	time.Sleep(5 * time.Second)
	if exchange.deleteActive == now {
		exchange.deleteChan <- exchange
	}
}
