package main

import (
	"bytes"
	"container/list"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
	"net"
	"strings"
	"sync"
)

type Server struct {
	exchanges  map[string]*Exchange
	queues     map[string]*Queue
	bindings   []*Binding
	idLock     sync.Mutex
	conns      map[int64]*AMQPConnection
	db         *bolt.DB
	serverLock sync.Mutex
	msgStore   *MessageStore
}

func (server *Server) MarshalJSON() ([]byte, error) {
	conns := make(map[string]*AMQPConnection)
	for id, value := range server.conns {
		conns[fmt.Sprintf("%d", id)] = value
	}
	return json.Marshal(map[string]interface{}{
		"exchanges":   server.exchanges,
		"queues":      server.queues,
		"connections": conns,
	})
}

func NewServer(dbPath string) *Server {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		panic(err.Error())

	}
	msgStore, err := NewMessageStore("msg_store.db")
	if err != nil {
		panic("Could not create message store!")
	}

	var server = &Server{
		exchanges: make(map[string]*Exchange),
		queues:    make(map[string]*Queue),
		bindings:  make([]*Binding, 0),
		conns:     make(map[int64]*AMQPConnection),
		db:        db,
		msgStore:  msgStore,
	}

	server.init()
	return server
}

func (server *Server) init() {
	server.initExchanges()
	server.initQueues()
	server.initBindings()
}

func (server *Server) initBindings() {
	fmt.Printf("Loading bindings from disk\n")
	// LOAD FROM PERSISTENT STORAGE
	err := server.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("bindings"))
		if bucket == nil {
			return nil
		}
		// iterate through queues
		cursor := bucket.Cursor()
		for name, data := cursor.First(); name != nil; name, data = cursor.Next() {
			// Read
			var method = &amqp.QueueBind{}
			var reader = bytes.NewReader(data)
			var err = method.Read(reader)
			if err != nil {
				panic(fmt.Sprintf("Failed to read binding '%s': %s", name, err.Error()))
			}

			// Get Exchange
			var exchange, foundExchange = server.exchanges[method.Exchange]
			if !foundExchange {
				return fmt.Errorf("Couldn't bind non-existant exchange %s", method.Exchange)
			}

			// Add Binding
			exchange.addBinding(method, -1, true)
		}
		return nil
	})
	if err != nil {
		panic("********** FAILED TO LOAD BINDINGS: " + err.Error())
	}
}

func (server *Server) initQueues() {
	fmt.Printf("Loading queues from disk\n")
	// LOAD FROM PERSISTENT STORAGE
	err := server.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("queues"))
		if bucket == nil {
			return nil
		}
		// iterate through queues
		cursor := bucket.Cursor()
		for name, data := cursor.First(); name != nil; name, data = cursor.Next() {
			var method = &amqp.QueueDeclare{}
			var reader = bytes.NewReader(data)
			var err = method.Read(reader)
			if err != nil {
				panic(fmt.Sprintf("Failed to read queue '%s': %s", name, err.Error()))
			}
			// fmt.Printf("Got queue from disk: %s\n", method.Queue)
			_, err = server.declareQueue(method, -1, true)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic("********** FAILED TO LOAD QUEUES: " + err.Error())
	}
}

func (server *Server) initExchanges() {
	fmt.Printf("Loading exchanges from disk\n")
	// LOAD FROM PERSISTENT STORAGE
	err := server.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("exchanges"))
		if bucket == nil {
			return nil
		}
		// iterate through exchanges
		cursor := bucket.Cursor()
		for name, data := cursor.First(); name != nil; name, data = cursor.Next() {
			// TODO: maybe convert the name ~ to empty string?
			var method = &amqp.ExchangeDeclare{}
			var reader = bytes.NewReader(data)
			var err = method.Read(reader)
			if err != nil {
				panic(fmt.Sprintf("Failed to read exchange '%s': %s", name, err.Error()))
			}
			system, err := amqp.ReadBit(reader)
			// fmt.Printf("Got exchange from disk: %s (%s)\n", method.Exchange, method.Type)
			_, err = server.declareExchange(method, system, true)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic("********** FAILED TO LOAD EXCHANGES: " + err.Error())
	}

	// DECLARE MISSING SYSEM EXCHANGES
	var _, hasKey = server.exchanges[""]
	if !hasKey {
		fmt.Println("System exchange '' doesn't exist. Creating")
		_, err := server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "",
			Type:       "direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.NewTable(),
		}, true, false)
		if err != nil {
			panic("Error making exchange: " + err.Error())
		}
	}

	// amq.direct
	_, hasKey = server.exchanges["amq.direct"]
	if !hasKey {
		fmt.Println("System exchange 'amq.direct' doesn't exist. Creating")
		_, err := server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "amq.direct",
			Type:       "direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.NewTable(),
		}, true, false)
		if err != nil {
			panic("Error making exchange: " + err.Error())
		}
	}

	// amqp.fanout
	_, hasKey = server.exchanges["amq.fanout"]
	if !hasKey {
		fmt.Println("System exchange 'amq.fanout' doesn't exist. Creating")
		_, err = server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "amq.fanout",
			Type:       "fanout",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.NewTable(),
		}, true, false)
		if err != nil {
			panic("Error making exchange: " + err.Error())
		}
	}
	// amqp.topic
	_, hasKey = server.exchanges["amq.topic"]
	if !hasKey {
		fmt.Println("System exchange 'amq.topic' doesn't exist. Creating")
		_, err = server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "amq.topic",
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.NewTable(),
		}, true, false)
		if err != nil {
			panic("Error making exchange: " + err.Error())
		}
	}
}

func (server *Server) declareExchange(method *amqp.ExchangeDeclare, system bool, diskLoad bool) (uint16, error) {
	var tp, err = exchangeNameToType(method.Type)
	if err != nil || tp == EX_TYPE_HEADERS {
		// TODO: I should really make ChannelException and ConnectionException
		// types
		return uint16(503), err
	}
	var exchange = &Exchange{
		name:       method.Exchange,
		extype:     tp,
		durable:    method.Durable,
		autodelete: method.AutoDelete,
		internal:   method.Internal,
		arguments:  method.Arguments,
		incoming:   make(chan amqp.Frame),
		bindings:   make([]*Binding, 0),
		system:     system,
		server:     server,
	}
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	existing, hasKey := server.exchanges[exchange.name]
	if !hasKey && method.Passive {
		return 404, errors.New("Exchange does not exist")
	}
	if hasKey {
		if diskLoad {
			panic(fmt.Sprintf("Can't disk load a key that exists: %s", exchange.name))
		}
		if existing.extype != exchange.extype {
			return 530, errors.New("Cannot redeclare an exchange with a different type")
		}
		if equivalentExchanges(existing, exchange) {
			return 0, nil
		}
		// Not equivalent, error in passive mode
		if method.Passive {
			return 406, errors.New("Exchange with this name already exists")
		}
	}
	if method.Passive {
		return 0, nil
	}

	// outside of passive mode you can't create an exchange starting with
	// amq.
	if strings.HasPrefix(method.Exchange, "amq.") && !system {
		return 0, errors.New("Exchange names starting with 'amq.' are reserved")
	}

	if exchange.durable && !diskLoad {
		server.persistExchange(method, exchange.system)
	}
	server.exchanges[exchange.name] = exchange
	return 0, nil
}

func (server *Server) persistExchange(method *amqp.ExchangeDeclare, system bool) {
	err := server.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("exchanges"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		var buffer = bytes.NewBuffer(make([]byte, 0, 50)) // TODO: don't I know the size?
		method.Write(buffer)
		amqp.WriteBit(buffer, system)
		var name = method.Exchange
		if name == "" {
			name = "~"
		}
		// trim off the first four bytes, they're the class/method, which we
		// already know
		return bucket.Put([]byte(name), buffer.Bytes()[4:])
	})
	if err != nil {
		fmt.Printf("********** FAILED TO PERSIST EXCHANGE '%s': %s\n", method.Exchange, err.Error())
	}
}

func (server *Server) persistQueue(method *amqp.QueueDeclare) {
	err := server.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("queues"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		var buffer = bytes.NewBuffer(make([]byte, 0, 50)) // TODO: don't I know the size?
		method.Write(buffer)
		// trim off the first four bytes, they're the class/method, which we
		// already know
		return bucket.Put([]byte(method.Queue), buffer.Bytes()[4:])
	})
	if err != nil {
		fmt.Printf("********** FAILED TO PERSIST QUEUE '%s': %s\n", method.Queue, err.Error())
	}
}

func (server *Server) persistBinding(method *amqp.QueueBind) error {
	return server.db.Update(func(tx *bolt.Tx) error {
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

		return bucket.Put([]byte(hash.Sum(nil)), value)
	})
}

func (server *Server) depersistBinding(binding *Binding) error {
	return server.db.Update(func(tx *bolt.Tx) error {
		return depersistBinding(tx, binding)
	})
}

func (server *Server) declareQueue(method *amqp.QueueDeclare, connId int64, fromDisk bool) (string, error) {
	if !method.Exclusive {
		connId = -1
	}
	var queue = &Queue{
		name:       method.Queue,
		durable:    method.Durable,
		exclusive:  method.Exclusive,
		autoDelete: method.AutoDelete,
		arguments:  method.Arguments,
		queue:      list.New(),
		consumers:  make([]*Consumer, 0, 1),
		maybeReady: make(chan bool, 1),
		connId:     connId,
		server:     server,
	}
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	_, hasKey := server.queues[queue.name]
	if hasKey {
		return queue.name, nil
	}
	server.queues[queue.name] = queue
	var defaultExchange = server.exchanges[""]
	var defaultBinding = &amqp.QueueBind{
		Queue:      queue.name,
		Exchange:   "",
		RoutingKey: queue.name,
		Arguments:  amqp.NewTable(),
	}
	defaultExchange.addBinding(defaultBinding, connId, fromDisk)
	// TODO: queue should store bindings too?
	if method.Durable && !fromDisk {
		server.persistQueue(method)
	}
	queue.start()
	return queue.name, nil
}

func (server *Server) deleteQueuesForConn(connId int64) {
	server.serverLock.Lock()
	var queues = make([]*Queue, 0)
	for _, queue := range server.queues {
		if queue.connId == connId {
			queues = append(queues, queue)
		}
	}
	server.serverLock.Unlock()
	for _, queue := range queues {
		var method = &amqp.QueueDelete{
			Queue: queue.name,
		}
		server.deleteQueue(method, connId)
	}
}

func (server *Server) deleteQueue(method *amqp.QueueDelete, connId int64) (uint32, uint16, error) {
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	// Validate
	var queue, foundQueue = server.queues[method.Queue]
	if !foundQueue {
		return 0, 404, errors.New("Queue not found")
	}

	if queue.connId != -1 && queue.connId != connId {
		return 0, 405, fmt.Errorf("Queue is locked to another connection")
	}

	// Close to stop anything from changing
	queue.close()
	// Delete for storage
	server.depersistQueue(queue)
	server.removeBindingsForQueue(method.Queue)
	// Cleanup
	numPurged, err := queue.delete(method.IfUnused, method.IfEmpty)
	delete(server.queues, method.Queue)
	if err != nil {
		return 0, 406, err
	}
	return numPurged, 0, nil

}

func (server *Server) depersistQueue(queue *Queue) error {
	bindings := server.bindingsForQueue(queue.name)
	return server.db.Update(func(tx *bolt.Tx) error {
		for _, binding := range bindings {
			if err := depersistBinding(tx, binding); err != nil {
				return err
			}
		}
		return depersistQueue(tx, queue)
	})
}

func (server *Server) depersistExchange(exchange *Exchange) error {
	return server.db.Update(func(tx *bolt.Tx) error {
		for _, binding := range exchange.bindings {
			if err := depersistBinding(tx, binding); err != nil {
				return err
			}
		}
		return depersistExchange(tx, exchange)
	})
}

func (server *Server) bindingsForQueue(queueName string) []*Binding {
	ret := make([]*Binding, 0)
	for _, exchange := range server.exchanges {
		ret = append(ret, exchange.bindingsForQueue(queueName)...)
	}
	return ret
}

func (server *Server) removeBindingsForQueue(queueName string) {
	for _, exchange := range server.exchanges {
		exchange.removeBindingsForQueue(queueName)
	}
}

func (server *Server) deleteExchange(method *amqp.ExchangeDelete) (uint16, error) {
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	exchange, found := server.exchanges[method.Exchange]
	if !found {
		return 404, fmt.Errorf("Exchange not found: '%s'", method.Exchange)
	}
	if exchange.system {
		return 530, fmt.Errorf("Cannot delete system exchange: '%s'", method.Exchange)
	}
	exchange.close()
	server.depersistExchange(exchange)
	// Note: we don't need to delete the bindings from the queues they are
	// associated with because they are stored on the exchange.
	delete(server.exchanges, method.Exchange)
	return 0, nil
}

func (server *Server) deregisterConnection(id int64) {
	delete(server.conns, id)
}

func (server *Server) openConnection(network net.Conn) {
	c := NewAMQPConnection(server, network)
	server.conns[c.id] = c
	c.openConnection()
}
