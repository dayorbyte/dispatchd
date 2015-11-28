package main

import (
	_ "fmt"
	"github.com/jeffjenkins/mq/amqp"
	"github.com/jeffjenkins/mq/exchange"
	"strings"
)

func (channel *Channel) exchangeRoute(methodFrame amqp.MethodFrame) *amqp.AMQPError {
	switch method := methodFrame.(type) {
	case *amqp.ExchangeDeclare:
		return channel.exchangeDeclare(method)
	case *amqp.ExchangeBind:
		return channel.exchangeBind(method)
	case *amqp.ExchangeUnbind:
		return channel.exchangeUnbind(method)
	case *amqp.ExchangeDelete:
		return channel.exchangeDelete(method)
	}
	var classId, methodId = methodFrame.MethodIdentifier()
	return amqp.NewHardError(540, "Not implemented", classId, methodId)
}

func (channel *Channel) exchangeDeclare(method *amqp.ExchangeDeclare) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	// The client I'm using for testing thought declaring the empty exchange
	// was OK. Check later
	// if len(method.Exchange) > 0 && !method.Passive {
	// 	var msg = "The empty exchange name is reserved"
	// 	channel.channelErrorWithMethod(406, msg, classId, methodId)
	// 	return nil
	// }

	// Check the name format
	var err = amqp.CheckExchangeOrQueueName(method.Exchange)
	if err != nil {
		return amqp.NewSoftError(406, err.Error(), classId, methodId)
	}

	// Declare!
	var ex, amqpErr = exchange.NewFromMethod(method, false, channel.server.exchangeDeleter)
	if amqpErr != nil {
		return amqpErr
	}
	tp, err := exchange.ExchangeNameToType(method.Type)
	if err != nil || tp == exchange.EX_TYPE_HEADERS {
		return amqp.NewHardError(503, err.Error(), classId, methodId)
	}
	existing, hasKey := channel.server.exchanges[ex.Name]
	if !hasKey && method.Passive {
		return amqp.NewSoftError(404, "Exchange does not exist", classId, methodId)
	}
	if hasKey {
		// if diskLoad {
		// 	panic(fmt.Sprintf("Can't disk load a key that exists: %s", ex.Name))
		// }
		if existing.ExType != ex.ExType {
			return amqp.NewHardError(530, "Cannot redeclare an exchange with a different type", classId, methodId)
		}
		if existing.EquivalentExchanges(ex) {
			if !method.NoWait {
				channel.SendMethod(&amqp.ExchangeDeclareOk{})
			}
			return nil
		}
		// Not equivalent, error in passive mode
		if method.Passive {
			return amqp.NewSoftError(406, "Exchange with this name already exists", classId, methodId)
		}
	}
	if method.Passive {
		if !method.NoWait {
			channel.SendMethod(&amqp.ExchangeDeclareOk{})
		}
		return nil
	}

	// outside of passive mode you can't create an exchange starting with
	// amq.
	if strings.HasPrefix(method.Exchange, "amq.") {
		return amqp.NewSoftError(403, "Exchange names starting with 'amq.' are reserved", classId, methodId)
	}

	err = channel.server.addExchange(ex)
	if err != nil {
		return amqp.NewSoftError(500, err.Error(), classId, methodId)
	}
	err = ex.Persist(channel.server.db)
	if err != nil {
		return amqp.NewSoftError(500, err.Error(), classId, methodId)
	}
	if !method.NoWait {
		channel.SendMethod(&amqp.ExchangeDeclareOk{})
	}
	return nil
}

func (channel *Channel) exchangeDelete(method *amqp.ExchangeDelete) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	var errCode, err = channel.server.deleteExchange(method)
	if err != nil {
		return amqp.NewSoftError(errCode, err.Error(), classId, methodId)
	}
	if !method.NoWait {
		channel.SendMethod(&amqp.ExchangeDeleteOk{})
	}
	return nil
}

func (channel *Channel) exchangeBind(method *amqp.ExchangeBind) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	return amqp.NewHardError(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.SendMethod(&amqp.ExchangeBindOk{})
	// }
}
func (channel *Channel) exchangeUnbind(method *amqp.ExchangeUnbind) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	return amqp.NewHardError(540, "Not implemented", classId, methodId)
	// if !method.NoWait {
	// 	channel.SendMethod(&amqp.ExchangeUnbindOk{})
	// }
}
