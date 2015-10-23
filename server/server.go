package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/boltdb/bolt"
	"net"
	"strings"
	"sync"
)

type Server struct {
	exchanges map[string]*Exchange
	queues    map[string]*Queue
	bindings  []*Binding
	mutex     sync.Mutex
	nextId    uint64
	conns     map[uint64]*AMQPConnection
	db 				*bolt.DB
}

func (server *Server) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"exchanges": server.exchanges,
		"queues":    server.queues,
	})
}

func NewServer(dbPath string) *Server {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		panic(err.Error())

	}
	// defer db.Close()

	var server = &Server{
		exchanges: make(map[string]*Exchange),
		queues:    make(map[string]*Queue),
		bindings:  make([]*Binding, 0),
		conns:     make(map[uint64]*AMQPConnection),
		db:				 db,
	}

	server.init()

	server.createSystemExchanges()

	return server
}

func (server *Server) init() {
	// Load exchanges
	err := server.db.View(func(tx *bolt.Tx) error {
    return nil
	})
	if err != nil {
		panic("Could not init server: " + err.Error())
	}
}

func (server *Server) createSystemExchanges() {
	// Default exchange
	var _, hasKey = server.exchanges[""]
	if !hasKey {
		server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "",
			Type:       "direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.Table{},
		}, true)
	}

	// amq.direct
	_, hasKey = server.exchanges["amq.direct"]
	if !hasKey {
		server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "amq.direct",
			Type:     	"direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.Table{},
		}, true)
	}

	// amqp.fanout
	_, hasKey = server.exchanges["amq.fanout"]
	if hasKey {
		server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "amq.fanout",
			Type:     	"fanout",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.Table{},
		}, true)
	}
	// amqp.fanout
	_, hasKey = server.exchanges["amq.topic"]
	if hasKey {
		server.declareExchange(&amqp.ExchangeDeclare{
			Exchange:   "amq.topic",
			Type:     	"topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Arguments:  amqp.Table{},
		}, true)
	}
}

func (server *Server) declareExchange(method *amqp.ExchangeDeclare, system bool) (uint16, error) {
	var tp, err = exchangeNameToType(method.Type)
	if err != nil || tp == EX_TYPE_HEADERS {
		// TODO: I should really make ChannelException and ConnectionException
		// types
		return uint16(503), fmt.Errorf("Exchange type not supported: %s", method.Type)
	}
	var exchange = &Exchange{
		name:       method.Exchange,
		extype:     tp,
		durable:    method.Durable,
		autodelete: method.AutoDelete,
		internal:   method.Internal,
		arguments:  method.Arguments,
		incoming:   make(chan amqp.Frame),
		system:     false,
	}
	existing, hasKey := server.exchanges[exchange.name]
	if !hasKey && method.Passive {
		return 404, errors.New("Exchange does not exist")
	}
	if hasKey {
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
	if strings.HasPrefix(method.Exchange, "amq.") {
		return 0, errors.New("Exchange names starting with 'amq.' are reserved")
	}

	if exchange.durable {
		server.persistExchange(method)
	}
	server.exchanges[exchange.name] = exchange
	exchange.start()
	return 0, nil
}

func (server *Server) persistExchange(method *amqp.ExchangeDeclare) {

}

func (server *Server) declareQueue(method *amqp.QueueDeclare) (string, error) {
	var queue = &Queue{
		name:       method.Queue,
		durable:    method.Durable,
		exclusive:  method.Exclusive,
		autoDelete: method.AutoDelete,
		arguments:  method.Arguments,
		queue:      list.New(),
		consumers:  make([]*Consumer, 0, 1),
		maybeReady: make(chan bool, 1),
	}

	_, hasKey := server.queues[queue.name]
	if hasKey {
		return queue.name, nil
	}
	server.queues[queue.name] = queue
	var defaultExchange = server.exchanges[""]
	var defaultBinding = NewBinding(queue.name, "", queue.name, make(amqp.Table))
	defaultExchange.addBinding(queue, defaultBinding)
	// TODO: queue should store bindings too?
	queue.start()
	return queue.name, nil
}

func (server *Server) deleteQueue(method *amqp.QueueDelete) (uint32, uint16, error) {
	var queue, foundQueue = server.queues[method.Queue]
	if !foundQueue {
		return 0, 404, errors.New("Queue not found")
	}
	numPurged, err := queue.delete(method.IfUnused, method.IfEmpty)
	if err != nil {
		return 0, 406, err
	}
	return numPurged, 0, nil

}

func (server *Server) deleteExchange(method *amqp.ExchangeDelete) error {
	// TODO: clean up exchange
	delete(server.exchanges, method.Exchange)
	return nil
}

func (server *Server) nextConnId() uint64 {
	server.mutex.Lock()
	var id = server.nextId
	server.nextId += 1
	server.mutex.Unlock()
	return id
}

func (server *Server) deregisterConnection(id uint64) {
	delete(server.conns, id)
}

func (server *Server) openConnection(network net.Conn) {
	c := NewAMQPConnection(server, network)
	server.conns[c.id] = c
	c.openConnection()
}
