package main

import (
	"container/list"
	"errors"
	"github.com/jeffjenkins/mq/amqp"
	"net"
	"sync"
)

type Server struct {
	exchanges map[string]*Exchange
	queues    map[string]*Queue
	bindings  []*Binding
	mutex     sync.Mutex
	nextId    uint64
	conns     map[uint64]*AMQPConnection
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
}

func (server *Server) declareExchange(method *amqp.ExchangeDeclare) error {
	// TODO: Handle Passive
	var tp, err = exchangeNameToType(method.Type)
	if err != nil {
		// TODO: handle invalid exchange types
		panic(err.Error())
		return nil
	}
	var exchange = Exchange{
		name:       method.Exchange,
		extype:     tp,
		durable:    method.Durable,
		autodelete: method.AutoDelete,
		internal:   method.Internal,
		arguments:  method.Arguments,
		incoming:   make(chan amqp.Frame),
		system:     false,
	}
	_, hasKey := server.exchanges[exchange.name]
	if hasKey {
		return errors.New("Exchange with this name already exists")
	}
	server.exchanges[exchange.name] = &exchange
	exchange.start()
	return nil
}

func (server *Server) declareQueue(method *amqp.QueueDeclare) error {
	var queue = &Queue{
		name:       method.Queue,
		durable:    method.Durable,
		exclusive:  method.Exclusive,
		autoDelete: method.AutoDelete,
		arguments:  method.Arguments,
		queue:      list.New(),
		consumers:  list.New(),
	}
	_, hasKey := server.queues[queue.name]
	if hasKey {
		// TODO(MUST): channel exception if there is a queue with the same name
		// and different properties
		return nil
	}
	server.queues[queue.name] = queue
	var defaultExchange = server.exchanges[""]
	var defaultBinding = &Binding{
		queueName:    queue.name,
		exchangeName: "",
		key:          queue.name,
		arguments:    make(amqp.Table),
	}
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
