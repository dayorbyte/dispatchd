package main

import (
	"github.com/jeffjenkins/dispatchd/amqp"
)

func (channel *Channel) txRoute(methodFrame amqp.MethodFrame) *amqp.AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.TxSelect:
		return channel.txSelect(method)
	case *amqp.TxCommit:
		return channel.txCommit(method)
	case *amqp.TxRollback:
		return channel.txRollback(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return amqp.NewHardError(540, "Unable to route method frame", classId, methodId)
}

func (channel *Channel) txSelect(method *amqp.TxSelect) *amqp.AMQPError {
	channel.startTxMode()
	channel.SendMethod(&amqp.TxSelectOk{})
	return nil
}

func (channel *Channel) txCommit(method *amqp.TxCommit) *amqp.AMQPError {
	if amqpErr := channel.commitTx(); amqpErr != nil {
		return amqpErr
	}
	channel.SendMethod(&amqp.TxCommitOk{})
	return nil
}

func (channel *Channel) txRollback(method *amqp.TxRollback) *amqp.AMQPError {
	channel.rollbackTx()
	channel.SendMethod(&amqp.TxRollbackOk{})
	return nil
}
