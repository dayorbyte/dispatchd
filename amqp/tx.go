package amqp

func NewTxMessage(msg *Message, queueName string) *TxMessage {
	return &TxMessage{
		QueueName: queueName,
		Msg:       msg,
	}
}
