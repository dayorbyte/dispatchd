package amqp

func NewTxMessage(msg *Message, queueName string) *TxMessage {
	return &TxMessage{
		QueueName: queueName,
		Msg:       msg,
	}
}

func NewTxAck(tag uint64, nack bool, requeueNack bool, multiple bool) *TxAck {
	return &TxAck{
		Tag:         tag,
		Nack:        nack,
		RequeueNack: requeueNack,
		Multiple:    multiple,
	}
}
