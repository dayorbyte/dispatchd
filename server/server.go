package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/binding"
	"github.com/jeffjenkins/mq/exchange"
	"github.com/jeffjenkins/mq/msgstore"
	"github.com/jeffjenkins/mq/queue"
	"net"
	"strings"
	"sync"
)

type Server struct {
	exchanges       map[string]*exchange.Exchange
	queues          map[string]*queue.Queue
	bindings        []*binding.Binding
	idLock          sync.Mutex
	conns           map[int64]*AMQPConnection
	db              *bolt.DB
	serverLock      sync.Mutex
	msgStore        *msgstore.MessageStore
	exchangeDeleter chan *exchange.Exchange
	queueDeleter    chan *queue.Queue
}

func (server *Server) MarshalJSON() ([]byte, error) {
	conns := make(map[string]*AMQPConnection)
	for id, value := range server.conns {
		conns[fmt.Sprintf("%d", id)] = value
	}
	return json.Marshal(map[string]interface{}{
		"exchanges":     server.exchanges,
		"queues":        server.queues,
		"connections":   conns,
		"msgCount":      server.msgStore.MessageCount(),
		"msgIndexCount": server.msgStore.IndexCount(),
	})
}

func NewServer(dbPath string) *Server {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		panic(err.Error())

	}
	msgStore, err := msgstore.NewMessageStore("msg_store.db")
	if err != nil {
		panic("Could not create message store!")
	}

	var server = &Server{
		exchanges:       make(map[string]*exchange.Exchange),
		queues:          make(map[string]*queue.Queue),
		bindings:        make([]*binding.Binding, 0),
		conns:           make(map[int64]*AMQPConnection),
		db:              db,
		msgStore:        msgStore,
		exchangeDeleter: make(chan *exchange.Exchange),
		queueDeleter:    make(chan *queue.Queue),
	}

	server.init()
	return server
}

func (server *Server) init() {
	server.initExchanges()
	server.initQueues()
	server.initBindings()
	go server.exchangeDeleteMonitor()
	go server.queueDeleteMonitor()
}

func (server *Server) exchangeDeleteMonitor() {
	for e := range server.exchangeDeleter {
		var dele = &amqp.ExchangeDelete{
			Exchange: e.Name,
			NoWait:   true,
		}
		server.deleteExchange(dele)
	}
}

func (server *Server) queueDeleteMonitor() {
	for q := range server.queueDeleter {
		var delq = &amqp.QueueDelete{
			Queue:  q.Name,
			NoWait: true,
		}
		server.deleteQueue(delq, -1)
	}
}

// TODO: move most of this into the bindings file
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
			exchange.AddBinding(method, -1, true)
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
			fmt.Printf("Got queue from disk: %s\n", method.Queue)
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
	for name, queue := range server.queues {
		fmt.Printf("Loading persistent messages for queue '%s'\n", name)
		queue.LoadFromMsgStore(server.msgStore)

		fmt.Printf("Loaded %d messages from queue\n", queue.Len())
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
	var tp, err = exchange.ExchangeNameToType(method.Type)
	if err != nil || tp == exchange.EX_TYPE_HEADERS {
		// TODO: I should really make ChannelException and ConnectionException
		// types
		return uint16(503), err
	}
	var exchange = exchange.NewExchange(
		method.Exchange,
		tp,
		method.Durable,
		method.AutoDelete,
		method.Internal,
		method.Arguments,
		system,
		server.exchangeDeleter,
	)
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	existing, hasKey := server.exchanges[exchange.Name]
	if !hasKey && method.Passive {
		return 404, errors.New("Exchange does not exist")
	}
	if hasKey {
		if diskLoad {
			panic(fmt.Sprintf("Can't disk load a key that exists: %s", exchange.Name))
		}
		if existing.Extype != exchange.Extype {
			return 530, errors.New("Cannot redeclare an exchange with a different type")
		}
		if existing.EquivalentExchanges(exchange) {
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

	if exchange.Durable && !diskLoad {
		server.persistExchange(method, exchange.System)
	}
	server.exchanges[exchange.Name] = exchange
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

func (server *Server) declareQueue(method *amqp.QueueDeclare, connId int64, fromDisk bool) (string, error) {
	if !method.Exclusive {
		connId = -1
	}
	var queue = queue.NewQueue(
		method.Queue,
		method.Durable,
		method.Exclusive,
		method.AutoDelete,
		method.Arguments,
		connId,
		server.msgStore,
		server.queueDeleter,
	)
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	_, hasKey := server.queues[queue.Name]
	if hasKey {
		return queue.Name, nil
	}
	server.queues[queue.Name] = queue
	var defaultExchange = server.exchanges[""]
	var defaultBinding = &amqp.QueueBind{
		Queue:      queue.Name,
		Exchange:   "",
		RoutingKey: queue.Name,
		Arguments:  amqp.NewTable(),
	}
	defaultExchange.AddBinding(defaultBinding, connId, fromDisk)
	// TODO: queue should store bindings too?
	if method.Durable && !fromDisk {
		server.persistQueue(method)
	}
	queue.Start()
	return queue.Name, nil
}

func (server *Server) deleteQueuesForConn(connId int64) {
	server.serverLock.Lock()
	var queues = make([]*queue.Queue, 0)
	for _, queue := range server.queues {
		if queue.ConnId == connId {
			queues = append(queues, queue)
		}
	}
	server.serverLock.Unlock()
	for _, queue := range queues {
		var method = &amqp.QueueDelete{
			Queue: queue.Name,
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

	if queue.ConnId != -1 && queue.ConnId != connId {
		return 0, 405, fmt.Errorf("Queue is locked to another connection")
	}

	// Close to stop anything from changing
	queue.Close()
	// Delete for storage
	server.depersistQueue(queue)
	server.removeBindingsForQueue(method.Queue)
	// Cleanup
	numPurged, err := queue.Delete(method.IfUnused, method.IfEmpty)
	delete(server.queues, method.Queue)
	if err != nil {
		return 0, 406, err
	}
	return numPurged, 0, nil

}

func (server *Server) depersistQueue(queue *queue.Queue) error {
	bindings := server.bindingsForQueue(queue.Name)
	return server.db.Update(func(tx *bolt.Tx) error {
		for _, binding := range bindings {
			if err := binding.DepersistBoltTx(tx); err != nil {
				return err
			}
		}
		return queue.DepersistBoltTx(tx)
	})
}

func (server *Server) bindingsForQueue(queueName string) []*binding.Binding {
	ret := make([]*binding.Binding, 0)
	for _, exchange := range server.exchanges {
		ret = append(ret, exchange.BindingsForQueue(queueName)...)
	}
	return ret
}

func (server *Server) removeBindingsForQueue(queueName string) {
	for _, exchange := range server.exchanges {
		exchange.RemoveBindingsForQueue(queueName)
	}
}

func (server *Server) deleteExchange(method *amqp.ExchangeDelete) (uint16, error) {
	server.serverLock.Lock()
	defer server.serverLock.Unlock()
	exchange, found := server.exchanges[method.Exchange]
	if !found {
		return 404, fmt.Errorf("Exchange not found: '%s'", method.Exchange)
	}
	if exchange.System {
		return 530, fmt.Errorf("Cannot delete system exchange: '%s'", method.Exchange)
	}
	exchange.Close()
	exchange.Depersist(server.db)
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

func (server *Server) returnMessage(msg *amqp.Message, code uint16, text string) *amqp.BasicReturn {
	return &amqp.BasicReturn{
		Exchange:   msg.Method.Exchange,
		RoutingKey: msg.Method.RoutingKey,
		ReplyCode:  code,
		ReplyText:  text,
	}
}

func (server *Server) publish(exchange *exchange.Exchange, msg *amqp.Message) (*amqp.BasicReturn, *AMQPError) {
	// Concurrency note: Since there is no lock we can, technically, have messages
	// published after the exchange has been closed. These couldn't be on the same
	// channel as the close is happening on, so that seems justifiable.
	if exchange.Closed {
		if msg.Method.Mandatory || msg.Method.Immediate {
			var rm = server.returnMessage(msg, 313, "Exchange closed, cannot route to queues or consumers")
			return rm, nil
		}
		return nil, nil
	}
	queues := exchange.QueuesForPublish(msg)

	if len(queues) == 0 {
		// If we got here the message was unroutable.
		if msg.Method.Mandatory || msg.Method.Immediate {
			var rm = server.returnMessage(msg, 313, "No queues available")
			return rm, nil
		}
	}

	var queueNames = make([]string, 0, len(queues))
	for k, _ := range queues {
		queueNames = append(queueNames, k)
	}

	// Immediate messages
	if msg.Method.Immediate {
		var consumed = false
		// Add message to message store
		queueMessagesByQueue, err := server.msgStore.AddMessage(msg, queueNames)
		if err != nil {
			return nil, NewSoftError(500, err.Error(), 60, 40)
		}
		// Try to immediately consumed it
		for queueName, _ := range queues {
			qms := queueMessagesByQueue[queueName]
			for _, qm := range qms {
				queue, found := server.queues[queueName]
				if !found {
					// The queue must have been deleted since the queuesForPublish call
					continue
				}
				var oneConsumed = queue.ConsumeImmediate(qm)
				var rhs = make([]amqp.MessageResourceHolder, 0)
				if !oneConsumed {
					server.msgStore.RemoveRef(qm, queueName, rhs)
				}
				consumed = oneConsumed || consumed
			}
		}
		if !consumed {
			var rm = server.returnMessage(msg, 313, "No consumers available for immediate message")
			return rm, nil
		}
		return nil, nil
	}

	// Add the message to the message store along with the queues we're about to add it to
	queueMessagesByQueue, err := server.msgStore.AddMessage(msg, queueNames)
	if err != nil {
		return nil, NewSoftError(500, err.Error(), 60, 40)
	}

	for queueName, _ := range queues {
		qms := queueMessagesByQueue[queueName]
		for _, qm := range qms {
			queue, found := server.queues[queueName]
			if !found || !queue.Add(qm) {
				// If we couldn't add it means the queue is closed and we should
				// remove the ref from the message store. The queue being closed means
				// it is going away, so worst case if the server dies we have to process
				// and discard the message on boot.
				var rhs = make([]amqp.MessageResourceHolder, 0)
				server.msgStore.RemoveRef(qm, queueName, rhs)
			}
		}
	}
	return nil, nil
}
