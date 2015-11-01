package main

import (
	"errors"
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) txRoute(methodFrame amqp.MethodFrame) error {
	switch method := methodFrame.(type) {
	case *amqp.TxSelect:
		return channel.txSelect(method)
	case *amqp.TxCommit:
		return channel.txCommit(method)
	case *amqp.TxRollback:
		return channel.txRollback(method)
	}
	return errors.New("Unable to route method frame")
}

func (channel *Channel) txSelect(method *amqp.TxSelect) error {
	channel.startTxMode()
	channel.sendMethod(&amqp.TxSelectOk{})
	return nil
}

func (channel *Channel) txCommit(method *amqp.TxCommit) error {
	channel.sendMethod(&amqp.TxCommitOk{})
	return nil
}

func (channel *Channel) txRollback(method *amqp.TxRollback) error {
	channel.sendMethod(&amqp.TxRollbackOk{})
	return nil
}
