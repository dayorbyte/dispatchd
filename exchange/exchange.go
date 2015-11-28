package exchange

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/binding"
	"github.com/jeffjenkins/mq/gen"
	"sync"
	"time"
)

const (
	EX_TYPE_DIRECT  uint8 = 1
	EX_TYPE_FANOUT  uint8 = 2
	EX_TYPE_TOPIC   uint8 = 3
	EX_TYPE_HEADERS uint8 = 4
)

type Exchange struct {
	gen.ExchangeState
	bindings         []*binding.Binding
	bindingsLock     sync.Mutex
	incoming         chan amqp.Frame
	Closed           bool
	deleteActive     time.Time
	deleteChan       chan *Exchange
	autodeletePeriod time.Duration
}

func (exchange *Exchange) Close() {
	exchange.Closed = true
}

func (exchange *Exchange) MarshalJSON() ([]byte, error) {
	var typ, err = exchangeTypeToName(exchange.ExType)
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
	extype uint8,
	durable bool,
	autodelete bool,
	internal bool,
	arguments *amqp.Table,
	system bool,
	deleteChan chan *Exchange,
) *Exchange {
	return &Exchange{
		ExchangeState: gen.ExchangeState{
			Name:       name,
			ExType:     extype,
			Durable:    durable,
			AutoDelete: autodelete,
			Internal:   internal,
			Arguments:  arguments,
			System:     system,
		},
		deleteChan: deleteChan,
		// not passed in
		incoming:         make(chan amqp.Frame),
		bindings:         make([]*binding.Binding, 0),
		autodeletePeriod: 5 * time.Second,
	}
}

func NewFromExchangeState(exState gen.ExchangeState, deleteChan chan *Exchange) *Exchange {
	return &Exchange{
		ExchangeState:    exState,
		deleteChan:       deleteChan,
		incoming:         make(chan amqp.Frame),
		bindings:         make([]*binding.Binding, 0),
		autodeletePeriod: 5 * time.Second,
	}
}

func NewFromMethod(method *amqp.ExchangeDeclare, system bool, exchangeDeleter chan *Exchange) (*Exchange, *amqp.AMQPError) {
	var classId, methodId = method.MethodIdentifier()
	var tp, err = ExchangeNameToType(method.Type)
	if err != nil || tp == EX_TYPE_HEADERS {
		// TODO: I should really make ChannelException and ConnectionException
		// types
		return nil, amqp.NewHardError(503, "Bad exchange type", classId, methodId)
	}
	var ex = NewExchange(
		method.Exchange,
		tp,
		method.Durable,
		method.AutoDelete,
		method.Internal,
		method.Arguments,
		system,
		exchangeDeleter,
	)
	return ex, nil
}

func (ex1 *Exchange) EquivalentExchanges(ex2 *Exchange) bool {
	// NOTE: auto-delete is ignored for existing exchanges, so we
	// do not check it here.
	if ex1.Name != ex2.Name {
		return false
	}
	if ex1.ExType != ex2.ExType {
		return false
	}
	if ex1.Durable != ex2.Durable {
		return false
	}
	if ex1.Internal != ex2.Internal {
		return false
	}
	if !amqp.EquivalentTables(ex1.Arguments, ex2.Arguments) {
		return false
	}
	return true
}

func ExchangeNameToType(et string) (uint8, error) {
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

func exchangeTypeToName(et uint8) (string, error) {
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
	if msg.Method.Exchange != exchange.Name {
		return queues, nil
	}
	switch {
	case exchange.ExType == EX_TYPE_DIRECT:
		// In a direct exchange we can return the first match since there is
		// only one queue with a particular name
		for _, binding := range exchange.bindings {
			if binding.MatchDirect(msg.Method) {
				queues[binding.QueueName] = true
				return queues, nil
			}
		}
	case exchange.ExType == EX_TYPE_FANOUT:
		for _, binding := range exchange.bindings {
			queues[binding.QueueName] = true
		}
	case exchange.ExType == EX_TYPE_TOPIC:
		for _, binding := range exchange.bindings {
			if binding.MatchTopic(msg.Method) {
				var _, alreadySeen = queues[binding.QueueName]
				if alreadySeen {
					continue
				}
				queues[binding.QueueName] = true
			}
		}
	// case exchange.ExType == EX_TYPE_HEADERS:
	// 	// TODO: implement
	// 	panic("Headers is not implemented!")
	default: // pragma: nocover
		panic("Unknown exchange type created somehow. Server integrity error!")
	}
	return queues, nil
}

func (exchange *Exchange) Persist(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("exchanges"))
		if err != nil { // pragma: nocover
			return fmt.Errorf("create bucket: %s", err)
		}
		exBytes, err := exchange.ExchangeState.Marshal()
		if err != nil {
			return fmt.Errorf("Could not marshal exchange %s", exchange.Name)
		}

		var name = exchange.Name
		if name == "" {
			name = "~"
		}
		// trim off the first four bytes, they're the class/method, which we
		// already know
		return bucket.Put([]byte(name), exBytes)
	})
}

func NewFromDisk(db *bolt.DB, key string, deleteChan chan *Exchange) (ex *Exchange, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("exchanges"))
		if bucket == nil {
			return fmt.Errorf("Bucket not found: 'exchanges'")
		}
		ex, err = NewFromDiskBoltTx(bucket, []byte(key), deleteChan)
		return err
	})
	return
}

func NewFromDiskBoltTx(bucket *bolt.Bucket, key []byte, deleteChan chan *Exchange) (ex *Exchange, err error) {
	var lookupKey = key
	if len(key) == 0 {
		lookupKey = []byte{'~'}
	}
	exBytes := bucket.Get(lookupKey)
	if exBytes == nil {
		return nil, fmt.Errorf("Key not found: '%s'", key)
	}
	exState := gen.ExchangeState{}
	err = exState.Unmarshal(exBytes)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal exchange %s", key)
	}
	return NewFromExchangeState(exState, deleteChan), nil
}

func (exchange *Exchange) Depersist(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, binding := range exchange.bindings {
			if err := binding.DepersistBoltTx(tx); err != nil { // pragma: nocover
				return err
			}
		}
		return DepersistExchangeBoltTx(tx, exchange)
	})
}

func DepersistExchangeBoltTx(tx *bolt.Tx, exchange *Exchange) error {
	bucket, err := tx.CreateBucketIfNotExists([]byte("exchanges"))
	if err != nil { // pragma: nocover
		return fmt.Errorf("create bucket: %s", err)
	}
	return bucket.Delete([]byte(exchange.Name))
}

func (exchange *Exchange) IsTopic() bool {
	return exchange.ExType == EX_TYPE_TOPIC
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

	if exchange.AutoDelete {
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

func (exchange *Exchange) RemoveBinding(binding *binding.Binding) error {
	exchange.bindingsLock.Lock()
	defer exchange.bindingsLock.Unlock()

	// Delete binding
	for i, b := range exchange.bindings {
		if binding.Equals(b) {
			exchange.bindings = append(exchange.bindings[:i], exchange.bindings[i+1:]...)
			if exchange.AutoDelete && len(exchange.bindings) == 0 {
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
	time.Sleep(exchange.autodeletePeriod)
	if exchange.deleteActive == now {
		exchange.deleteChan <- exchange
	}
}
