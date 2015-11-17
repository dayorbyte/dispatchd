package main

import (
	"github.com/jeffjenkins/mq/amqp"
)

func (channel *Channel) txRoute(methodFrame amqp.MethodFrame) *AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.TxSelect:
		return channel.txSelect(method)
	case *amqp.TxCommit:
		return channel.txCommit(method)
	case *amqp.TxRollback:
		return channel.txRollback(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) txSelect(method *amqp.TxSelect) *AMQPError {
	channel.startTxMode()
	channel.sendMethod(&amqp.TxSelectOk{})
	return nil
}

func (channel *Channel) txCommit(method *amqp.TxCommit) *AMQPError {
	if amqpErr := channel.commitTx(); amqpErr != nil {
		return amqpErr
	}
	channel.sendMethod(&amqp.TxCommitOk{})
	return nil
}

func (channel *Channel) txRollback(method *amqp.TxRollback) *AMQPError {
	channel.rollbackTx()
	channel.sendMethod(&amqp.TxRollbackOk{})
	return nil
}
