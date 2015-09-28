package main

import (
	"errors"
	"github.com/jeffjenkins/mq/amqp"
)

type Server struct {
	exchanges map[string]*Exchange
	queues    map[string]*Queue
	bindings  []*Binding
}

func NewServer() *Server {
	return &Server{make(map[string]*Exchange), make(map[string]*Queue), []*Binding{}}
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
