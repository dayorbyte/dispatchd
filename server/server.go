package main

import (
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
	return &Server{
		exchanges: make(map[string]*Exchange),
		queues:    make(map[string]*Queue),
		bindings:  make([]*Binding, 0),
		conns:     make(map[uint64]*AMQPConnection),
	}
}

func (server *Server) declareExchange(method *amqp.ExchangeDeclare) error {
	var exchange = Exchange{
		method.Exchange,
		method.Type,
		method.Passive,
		method.Durable,
		method.AutoDelete,
		method.Internal,
		method.NoWait,
		method.Arguments,
		make(chan amqp.Frame),
	}
	_, hasKey := server.exchanges[exchange.name]
	if hasKey {
		return errors.New("Exchange with this name already exists")
	}
	server.exchanges[exchange.name] = &exchange
	exchange.start()
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
