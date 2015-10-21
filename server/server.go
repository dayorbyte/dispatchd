package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
	"net"
	"sync"
	"strings"
)

type Server struct {
	exchanges map[string]*Exchange
	queues    map[string]*Queue
	bindings  []*Binding
	mutex     sync.Mutex
	nextId    uint64
	conns     map[uint64]*AMQPConnection
}

func (server *Server) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"exchanges": server.exchanges,
		"queues":    server.queues,
	})
}

func NewServer() *Server {
	var server = &Server{
		exchanges: make(map[string]*Exchange),
		queues:    make(map[string]*Queue),
		bindings:  make([]*Binding, 0),
		conns:     make(map[uint64]*AMQPConnection),
	}

	server.createSystemExchanges()

	return server
}

func (server *Server) createSystemExchanges() {
	// Default exchange
	var defaultEx = &Exchange{
		name:       "",
		extype:     EX_TYPE_DIRECT,
		durable:    false,
		autodelete: false,
		internal:   false,
		arguments:  amqp.Table{},
		incoming:   make(chan amqp.Frame),
		system:     true,
	}
	var _, hasKey = server.exchanges[defaultEx.name]
	if hasKey {
		panic("Default system exchange already exists!")
	}
	server.exchanges[defaultEx.name] = defaultEx

	// amq.direct
	var directEx = &Exchange{
		name:       "amq.direct",
		extype:     EX_TYPE_DIRECT,
		durable:    false,
		autodelete: false,
		internal:   false,
		arguments:  amqp.Table{},
		incoming:   make(chan amqp.Frame),
		system:     true,
	}
	_, hasKey = server.exchanges[directEx.name]
	if hasKey {
		panic("amq.direct system exchange already exists!")
	}
	server.exchanges[directEx.name] = directEx

	// amqp.fanout
	var fanoutEx = &Exchange{
		name:       "amq.fanout",
		extype:     EX_TYPE_FANOUT,
		durable:    false,
		autodelete: false,
		internal:   false,
		arguments:  amqp.Table{},
		incoming:   make(chan amqp.Frame),
		system:     true,
	}
	_, hasKey = server.exchanges[fanoutEx.name]
	if hasKey {
		panic("amq.fanout system exchange already exists!")
	}
	server.exchanges[fanoutEx.name] = fanoutEx

	// amqp.fanout
	var topicEx = &Exchange{
		name:       "amq.topic",
		extype:     EX_TYPE_TOPIC,
		durable:    false,
		autodelete: false,
		internal:   false,
		arguments:  amqp.Table{},
		incoming:   make(chan amqp.Frame),
		system:     true,
	}
	_, hasKey = server.exchanges[topicEx.name]
	if hasKey {
		panic("amq.fanout system exchange already exists!")
	}
	server.exchanges[topicEx.name] = topicEx
}

func (server *Server) declareExchange(method *amqp.ExchangeDeclare) (uint16, error) {
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


	server.exchanges[exchange.name] = exchange
	exchange.start()
	return 0, nil
}

func (server *Server) declareQueue(method *amqp.QueueDeclare) error {
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
		// TODO(MUST): channel exception if there is a queue with the same name
		// and different properties
		return nil
	}
	server.queues[queue.name] = queue
	var defaultExchange = server.exchanges[""]
	var defaultBinding = NewBinding(queue.name, "", queue.name, make(amqp.Table))
	defaultExchange.addBinding(queue, defaultBinding)
	// TODO: queue should store bindings too?
	queue.start()
	return nil
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
