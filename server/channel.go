package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jeffjenkins/dispatchd/amqp"
	"github.com/jeffjenkins/dispatchd/consumer"
	"github.com/jeffjenkins/dispatchd/queue"
	"github.com/jeffjenkins/dispatchd/stats"
	"math"
	"sync"
)

const (
	CH_STATE_INIT    = iota
	CH_STATE_OPEN    = iota
	CH_STATE_CLOSING = iota
	CH_STATE_CLOSED  = iota
)

type Channel struct {
	id             uint16
	server         *Server
	incoming       chan *amqp.WireFrame
	outgoing       chan *amqp.WireFrame
	conn           *AMQPConnection
	state          uint8
	confirmMode    bool
	currentMessage *amqp.Message
	consumers      map[string]*consumer.Consumer
	consumerLock   sync.Mutex
	sendLock       sync.Mutex
	lastQueueName  string
	flow           bool
	txMode         bool
	txLock         sync.Mutex
	txMessages     []*amqp.TxMessage
	txAcks         []*amqp.TxAck
	// Consumers
	msgIndex uint64
	// Delivery Tracking
	deliveryTag  uint64
	deliveryLock sync.Mutex
	ackLock      sync.Mutex
	awaitingAcks map[uint64]amqp.UnackedMessage
	// Channel QOS Limits
	limitLock     sync.Mutex
	prefetchSize  uint32
	prefetchCount uint16
	activeSize    uint32
	activeCount   uint16
	// Consumer default QOS limits
	defaultPrefetchSize  uint32
	defaultPrefetchCount uint16
	// Stats
	statPublish    stats.Histogram
	statRoute      stats.Histogram
	statSendChan   stats.Histogram
	statSendEncode stats.Histogram
}

func NewChannel(id uint16, conn *AMQPConnection) *Channel {
	// Perf note: The server is significantly more performant if there's a
	// buffer for incoming, but until there are metrics available to see
	// what in particular is slow I'm leaving it as-is
	return &Channel{
		id:           id,
		server:       conn.server,
		incoming:     make(chan *amqp.WireFrame, 100),
		outgoing:     conn.outgoing,
		conn:         conn,
		flow:         true,
		state:        CH_STATE_INIT,
		txMessages:   make([]*amqp.TxMessage, 0),
		txAcks:       make([]*amqp.TxAck, 0),
		consumers:    make(map[string]*consumer.Consumer),
		awaitingAcks: make(map[uint64]amqp.UnackedMessage),
		// Stats
		statPublish:    stats.MakeHistogram("statPublish"),
		statRoute:      stats.MakeHistogram("statRoute"),
		statSendChan:   stats.MakeHistogram("statSendChan"),
		statSendEncode: stats.MakeHistogram("statSendEncode"),
	}
}

func (channel *Channel) commitTx() *amqp.AMQPError {
	channel.txLock.Lock()
	defer channel.txLock.Unlock()
	// messages
	queueMessagesByQueue, err := channel.server.msgStore.AddTxMessages(channel.txMessages)
	if err != nil {
		return amqp.NewSoftError(500, err.Error(), 60, 40)
	}

	for queueName, qms := range queueMessagesByQueue {
		queue, found := channel.server.queues[queueName]
		// the
		if !found {
			continue
		}
		for _, qm := range qms {
			if !queue.Add(qm) {
				// If we couldn't add it means the queue is closed and we should
				// remove the ref from the message store. The queue being closed means
				// it is going away, so worst case if the server dies we have to process
				// and discard the message on boot.
				var rhs = []amqp.MessageResourceHolder{channel}
				channel.server.msgStore.RemoveRef(qm, queueName, rhs)
			}
		}
	}

	// Acks
	// todo: remove acked messages from persistent storage in a single
	// transaction
	for _, ack := range channel.txAcks {
		channel.txAckMessage(ack)
	}

	// Clear transaction
	channel.txMessages = make([]*amqp.TxMessage, 0)
	channel.txAcks = make([]*amqp.TxAck, 0)
	return nil
}

func (channel *Channel) txAckMessage(ack *amqp.TxAck) {
	switch {
	case ack.Multiple && !ack.Nack:
		channel.ackBelow(ack.Tag, true)
	case ack.Multiple && ack.Nack:
		channel.nackBelow(ack.Tag, ack.RequeueNack, true)
	case !ack.Multiple && !ack.Nack:
		channel.ackOne(ack.Tag, true)
	case !ack.Multiple && ack.Nack:
		channel.nackOne(ack.Tag, ack.RequeueNack, true)
	}
}

func (channel *Channel) rollbackTx() {
	channel.txLock.Lock()
	defer channel.txLock.Unlock()
	channel.txMessages = make([]*amqp.TxMessage, 0)
	channel.txAcks = make([]*amqp.TxAck, 0)
}

func (channel *Channel) startTxMode() {
	channel.txMode = true
}

func (channel *Channel) recover(requeue bool) {
	if requeue {
		channel.ackLock.Lock()
		defer channel.ackLock.Unlock()
		// Requeue. Make sure we update stats
		for _, unacked := range channel.awaitingAcks {
			// re-add to queue
			queue, qFound := channel.server.queues[unacked.QueueName]
			if qFound {
				queue.Readd(unacked.QueueName, unacked.Msg)
			}
			// else: The queue gone. The reference would have been removed
			//       then so we don't remove it now in an else clause

			consumer, cFound := channel.consumers[unacked.ConsumerTag]
			// decr channel active
			channel.ReleaseResources(unacked.Msg)
			// decr consumer active
			if cFound {
				consumer.ReleaseResources(unacked.Msg)
			}
		}
		// Clear awaiting acks
		channel.awaitingAcks = make(map[uint64]amqp.UnackedMessage)
	} else {
		// Redeliver. Don't need to mess with stats.
		// We do this in a short-lived goroutine since this could end up
		// blocking on sending to the network inside the consumer
		go func() {
			for tag, unacked := range channel.awaitingAcks {
				consumer, cFound := channel.consumers[unacked.ConsumerTag]
				if cFound {
					// Consumer exists, try to deliver again
					channel.server.msgStore.IncrDeliveryCount(unacked.QueueName, unacked.Msg)
					consumer.Redeliver(tag, unacked.Msg)
				} else {
					// no consumer, drop message
					var rhs = []amqp.MessageResourceHolder{
						channel,
						consumer,
					}
					channel.server.msgStore.RemoveRef(unacked.Msg, unacked.QueueName, rhs)
				}

			}
		}()

	}
}

func (channel *Channel) changeFlow(active bool) {
	if channel.flow == active {
		return
	}
	channel.flow = active
	// If flow is active again, ping the consumers to let them try getting
	// work again.
	if channel.flow {
		for _, consumer := range channel.consumers {
			consumer.Ping()
		}
	}
}

func (channel *Channel) ackBelow(tag uint64, commitTx bool) *amqp.AMQPError {
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, amqp.NewTxAck(tag, false, false, true))
		return nil
	}
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	for k, unacked := range channel.awaitingAcks {
		if k <= tag || tag == 0 {
			consumer, cFound := channel.consumers[unacked.ConsumerTag]
			// Initialize resource holders array
			var rhs = []amqp.MessageResourceHolder{channel}
			if cFound {
				rhs = append(rhs, consumer)
			}
			err := channel.server.msgStore.RemoveRef(unacked.Msg, unacked.QueueName, rhs)
			// TODO: if this was an error do I still delete the ack we're waiting for?
			// The resources probably haven't been released.
			if err != nil {
				return amqp.NewSoftError(500, err.Error(), 60, 80)
			}
			delete(channel.awaitingAcks, k)

			if cFound {
				consumer.Ping()
			}
		}
	}
	// TODO: should this be an error if nothing was actually deleted and tag != 0?
	return nil
}

func (channel *Channel) ackOne(tag uint64, commitTx bool) *amqp.AMQPError {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()

	var unacked, found = channel.awaitingAcks[tag]
	if !found {
		var msg = fmt.Sprintf("Precondition Failed: Delivery Tag not found: %d", tag)
		return amqp.NewSoftError(406, msg, 60, 80)
	}
	// Tx mode
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, amqp.NewTxAck(tag, false, false, false))
		return nil
	}
	// Normal mode
	// Init
	consumer, cFound := channel.consumers[unacked.ConsumerTag]

	// Initialize resource holders array
	var rhs = []amqp.MessageResourceHolder{channel}
	if cFound {
		rhs = append(rhs, consumer)
	}
	err := channel.server.msgStore.RemoveRef(unacked.Msg, unacked.QueueName, rhs)
	// TODO: if this is an error, do I still delete the tag? the resources
	// probably haven't been freed
	if err != nil {
		return amqp.NewSoftError(500, err.Error(), 60, 80)
	}
	delete(channel.awaitingAcks, tag)

	if cFound {
		consumer.Ping()
	}
	return nil
}

func (channel *Channel) nackBelow(tag uint64, requeue bool, commitTx bool) *amqp.AMQPError {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()

	// Transaction mode
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, amqp.NewTxAck(tag, true, requeue, true))
		return nil
	}

	// Non-transaction mode
	var count = 0
	for k, unacked := range channel.awaitingAcks {
		if k <= tag || tag == 0 {
			count += 1
			// Init
			consumer, cFound := channel.consumers[unacked.ConsumerTag]
			queue, qFound := channel.server.queues[unacked.QueueName]

			// Initialize resource holders array
			var rhs = []amqp.MessageResourceHolder{channel}
			if cFound {
				rhs = append(rhs, consumer)
			}

			// requeue and release the approriate resources
			if requeue && qFound {
				// If we're requeueing we release the resources but don't remove the
				// reference.
				queue.Readd(unacked.QueueName, unacked.Msg)
				for _, rh := range rhs {
					rh.ReleaseResources(unacked.Msg)
				}
			} else {
				// If we aren't re-adding, remove the ref and all associated
				// resources
				err := channel.server.msgStore.RemoveRef(unacked.Msg, unacked.QueueName, rhs)
				if err != nil {
					return amqp.NewSoftError(500, err.Error(), 60, 120)
				}
			}

			// Remove this unacked message from the ones
			// we're waiting for acks on and ping the consumer
			// since there might be a message available now
			delete(channel.awaitingAcks, k)
			if cFound {
				consumer.Ping()
			}
		}
	}
	return nil
}

func (channel *Channel) nackOne(tag uint64, requeue bool, commitTx bool) *amqp.AMQPError {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	var unacked, found = channel.awaitingAcks[tag]
	if !found {
		var msg = fmt.Sprintf("Precondition Failed: Delivery Tag not found: %d", tag)
		return amqp.NewSoftError(406, msg, 60, 120)
	}
	// Transaction mode
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, amqp.NewTxAck(tag, true, requeue, false))
		return nil
	}
	// Non-transaction mode
	// Init
	consumer, cFound := channel.consumers[unacked.ConsumerTag]
	queue, qFound := channel.server.queues[unacked.QueueName]

	// Initialize resource holders array
	var rhs = []amqp.MessageResourceHolder{channel}
	if cFound {
		rhs = append(rhs, consumer)
	}

	// requeue and release the approriate resources
	if requeue && qFound {
		// If we're requeueing we release the resources but don't remove the
		// reference.
		queue.Readd(unacked.QueueName, unacked.Msg)
		for _, rh := range rhs {
			rh.ReleaseResources(unacked.Msg)
		}
	} else {
		// If we aren't re-adding, remove the ref and all associated
		// resources
		err := channel.server.msgStore.RemoveRef(unacked.Msg, unacked.QueueName, rhs)
		if err != nil {
			return amqp.NewSoftError(500, err.Error(), 60, 120)
		}
	}

	// Remove this unacked message from the ones
	// we're waiting for acks on and ping the consumer
	// since there might be a message available now
	delete(channel.awaitingAcks, tag)
	if cFound {
		consumer.Ping()
	}

	return nil
}

func (channel *Channel) FlowActive() bool {
	return channel.flow
}

func (channel *Channel) AddUnackedMessage(consumerTag string, msg *amqp.QueueMessage, queueName string) uint64 {
	var tag = channel.nextDeliveryTag()
	var unacked = amqp.NewUnackedMessage(consumerTag, msg, queueName)
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()

	_, found := channel.awaitingAcks[tag]
	if found {
		panic(fmt.Sprintf("Already found tag: %s", tag))
	}
	channel.awaitingAcks[tag] = *unacked
	// fmt.Printf("Adding tag: %d\n", tag)
	return tag
}

func (channel *Channel) addConsumer(q *queue.Queue, method *amqp.BasicConsume) *amqp.AMQPError {
	var classId, methodId = method.MethodIdentifier()
	// Create consumer
	var consumer = consumer.NewConsumer(
		channel.server.msgStore,
		method.Arguments,
		channel,
		method.ConsumerTag,
		method.Exclusive,
		method.NoAck,
		method.NoLocal,
		q,
		q.Name,
		channel.defaultPrefetchSize,
		channel.defaultPrefetchCount,
		channel.conn.id,
	)

	channel.consumerLock.Lock()
	defer channel.consumerLock.Unlock()
	// Make sure the doesn't exist on this channel
	_, found := channel.consumers[consumer.ConsumerTag]
	if found {
		return amqp.NewHardError(
			530,
			fmt.Sprintf("Consumer tag already exists: %s", consumer.ConsumerTag),
			classId,
			methodId,
		)
	}

	// Add the consumer to the queue, then channel
	code, err := q.AddConsumer(consumer, method.Exclusive)
	if err != nil {
		return amqp.NewSoftError(code, err.Error(), classId, methodId)
	}

	channel.consumers[consumer.ConsumerTag] = consumer
	consumer.Start()
	return nil
}

func (channel *Channel) ReleaseResources(qm *amqp.QueueMessage) {
	channel.limitLock.Lock()
	channel.activeCount -= 1
	channel.activeSize -= qm.MsgSize
	channel.limitLock.Unlock()
}

func (channel *Channel) AcquireResources(qm *amqp.QueueMessage) bool {
	channel.limitLock.Lock()
	defer channel.limitLock.Unlock()
	var sizeOk = channel.prefetchSize == 0 || channel.activeSize < channel.prefetchSize
	var countOk = channel.prefetchCount == 0 || channel.activeCount < channel.prefetchCount
	// If we're OK on size and count, acquire the resources
	if sizeOk && countOk {
		channel.activeCount += 1
		channel.activeSize += qm.MsgSize
		return true
	}
	return false
}

func (channel *Channel) setPrefetch(count uint16, size uint32, global bool) {
	if global {
		channel.prefetchSize = size
		channel.prefetchCount = count
	} else {
		channel.defaultPrefetchSize = size
		channel.defaultPrefetchCount = count
	}
}

func (channel *Channel) setStateOpen() {
	channel.state = CH_STATE_OPEN
}

func (channel *Channel) activateConfirmMode() {
	channel.confirmMode = true
}

func (channel *Channel) startPublish(method *amqp.BasicPublish) error {
	channel.currentMessage = amqp.NewMessage(method, channel.conn.id)
	return nil
}

func (channel *Channel) nextConfirmId() uint64 {
	channel.msgIndex++
	return channel.msgIndex
}

func (channel *Channel) nextDeliveryTag() uint64 {
	channel.deliveryLock.Lock()
	channel.deliveryTag++
	ret := channel.deliveryTag
	channel.deliveryLock.Unlock()
	return ret
}

func (channel *Channel) start() {
	if channel.id == 0 {
		channel.state = CH_STATE_OPEN
		go channel.startConnection()
	} else {
		go channel.startChannel()
	}

	// Receive method frames from the client and route them
	go func() {
		for {
			if channel.state == CH_STATE_CLOSED {
				break
			}
			var frame = <-channel.incoming
			var amqpErr *amqp.AMQPError = nil
			switch {
			case frame.FrameType == uint8(amqp.FrameMethod):
				amqpErr = channel.routeMethod(frame)
			case frame.FrameType == uint8(amqp.FrameHeader):
				if channel.state != CH_STATE_CLOSING {
					amqpErr = channel.handleContentHeader(frame)
				}
			case frame.FrameType == uint8(amqp.FrameBody):
				if channel.state != CH_STATE_CLOSING {
					amqpErr = channel.handleContentBody(frame)
				}
			default:
				amqpErr = amqp.NewHardError(500, "Unknown frame type", 0, 0)
			}
			if amqpErr != nil {
				channel.sendError(amqpErr)
			}
		}
	}()
}

func (channel *Channel) sendError(amqpErr *amqp.AMQPError) {
	if amqpErr.Soft {
		fmt.Println("Sending channel error:", amqpErr.Msg)
		channel.state = CH_STATE_CLOSING
		channel.SendMethod(&amqp.ChannelClose{
			ReplyCode: amqpErr.Code,
			ReplyText: amqpErr.Msg,
			ClassId:   amqpErr.Class,
			MethodId:  amqpErr.Method,
		})
	} else {
		channel.conn.connectionErrorWithMethod(amqpErr)
	}
}

func (channel *Channel) startChannel() {

}

func (channel *Channel) close(code uint16, reason string, classId uint16, methodId uint16) {
	channel.SendMethod(&amqp.ChannelClose{
		ReplyCode: code,
		ReplyText: reason,
		ClassId:   classId,
		MethodId:  methodId,
	})
	channel.state = CH_STATE_CLOSING
}

func (channel *Channel) shutdown() {
	if channel.state == CH_STATE_CLOSED {
		fmt.Printf("Shutdown already finished on %d\n", channel.id)
		return
	}
	channel.state = CH_STATE_CLOSED
	// unregister this channel
	channel.conn.deregisterChannel(channel.id)
	// remove any consumers associated with this channel
	for _, consumer := range channel.consumers {
		channel.removeConsumer(consumer.ConsumerTag)
	}
	// Any unacked messages should be re-added
	// for tag, unacked := range channel.awaitingAcks {
	// TODO(MUST): If we want at-most-once delivery we can't re-add these
	// messages. Need to figure out if the spec specifies, and after that
	// provide a way to have both options. Maybe a message header?
	// TODO(MUST): Is it safe to treat these as nacks?
	channel.nackBelow(math.MaxUint64, true, false)
}

func (channel *Channel) removeConsumer(consumerTag string) error {
	// TODO: how does this interact with the code in shutdown?
	var consumer, found = channel.consumers[consumerTag]
	if !found {
		return errors.New("Consumer not found")
	}
	consumer.Stop()
	delete(channel.consumers, consumerTag)
	return nil
}

// Send a method frame out to the client
// TODO: why isn't this taking a pointer?
func (channel *Channel) SendMethod(method amqp.MethodFrame) {
	// fmt.Printf("Sending method: %s\n", method.MethodName())
	var buf = bytes.NewBuffer([]byte{})
	method.Write(buf)
	channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameMethod), channel.id, buf.Bytes()}
}

// Send a method frame out to the client
func (channel *Channel) SendContent(method amqp.MethodFrame, message *amqp.Message) {
	var start = stats.Start()
	channel.sendLock.Lock()
	defer channel.sendLock.Unlock()
	// encode header
	var buf = bytes.NewBuffer(make([]byte, 0, 20)) // todo: don't I know the size?
	amqp.WriteShort(buf, message.Header.ContentClass)
	amqp.WriteShort(buf, message.Header.ContentWeight)
	amqp.WriteLonglong(buf, message.Header.ContentBodySize)
	var propBuf = bytes.NewBuffer(make([]byte, 0, 20))
	flags, err := message.Header.Properties.WriteProps(propBuf)
	if err != nil {
		panic("Error writing header!")
	}
	amqp.WriteShort(buf, flags)
	buf.Write(propBuf.Bytes())
	stats.RecordHisto(channel.statSendEncode, start)
	start = stats.Start()
	// Send method
	channel.SendMethod(method)
	// Send header
	channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameHeader), channel.id, buf.Bytes()}
	// Send body
	for _, b := range message.Payload {
		b.Channel = channel.id
		channel.outgoing <- b
	}
	stats.RecordHisto(channel.statSendChan, start)
}

func (channel *Channel) handleContentHeader(frame *amqp.WireFrame) *amqp.AMQPError {
	if channel.currentMessage == nil {
		return amqp.NewSoftError(500, "Unexpected content header frame!", 0, 0)
	}
	if channel.currentMessage.Header != nil {
		return amqp.NewSoftError(500, "Unexpected content header frame! Already saw header", 0, 0)
	}
	var headerFrame = &amqp.ContentHeaderFrame{}
	var err = headerFrame.Read(bytes.NewReader(frame.Payload))
	if err != nil {
		return amqp.NewHardError(500, "Error parsing header frame: "+err.Error(), 0, 0)
	}
	channel.currentMessage.Header = headerFrame
	return nil
}

func (channel *Channel) handleContentBody(frame *amqp.WireFrame) *amqp.AMQPError {
	if channel.currentMessage == nil {
		return amqp.NewSoftError(500, "Unexpected content body frame. No method content-having method called yet!", 0, 0)
	}
	if channel.currentMessage.Header == nil {
		return amqp.NewSoftError(500, "Unexpected content body frame! No header yet", 0, 0)
	}
	channel.currentMessage.Payload = append(channel.currentMessage.Payload, frame)
	// TODO: store this on message
	var size = uint64(0)
	for _, body := range channel.currentMessage.Payload {
		size += uint64(len(body.Payload))
	}
	if size < channel.currentMessage.Header.ContentBodySize {
		return nil
	}

	// We have the whole contents, let's publish!
	defer stats.RecordHisto(channel.statRoute, stats.Start())
	var server = channel.server
	var message = channel.currentMessage

	exchange, _ := server.exchanges[message.Method.Exchange]

	if channel.txMode {
		// TxMode, add the messages to a list
		queues, err := exchange.QueuesForPublish(channel.currentMessage)
		if err != nil {
			return err
		}

		channel.txLock.Lock()
		for queueName, _ := range queues {
			var txmsg = amqp.NewTxMessage(message, queueName)
			channel.txMessages = append(channel.txMessages, txmsg)
		}
		channel.txLock.Unlock()
	} else {
		// Normal mode, publish directly
		returnMethod, amqpErr := server.publish(exchange, channel.currentMessage)
		if amqpErr != nil {
			channel.currentMessage = nil
			return amqpErr
		}
		if returnMethod != nil {
			channel.SendContent(returnMethod, channel.currentMessage)
		}
	}

	channel.currentMessage = nil
	if channel.confirmMode {
		channel.msgIndex += 1
		channel.SendMethod(&amqp.BasicAck{channel.msgIndex, false})
	}
	return nil
}

func (channel *Channel) routeMethod(frame *amqp.WireFrame) *amqp.AMQPError {
	var methodReader = bytes.NewReader(frame.Payload)
	var methodFrame, err = amqp.ReadMethod(methodReader)
	if err != nil {
		return amqp.NewHardError(500, err.Error(), 0, 0)
	}
	var classId, methodId = methodFrame.MethodIdentifier()

	// If the method isn't closing related and we're closing, ignore the frames
	var closeChannel = classId == amqp.ClassIdChannel && (methodId == amqp.MethodIdChannelClose || methodId == amqp.MethodIdChannelCloseOk)
	var closeConnection = classId == amqp.ClassIdConnection && (methodId == amqp.MethodIdConnectionClose || methodId == amqp.MethodIdConnectionCloseOk)
	if channel.state == CH_STATE_CLOSING && !(closeChannel || closeConnection) {
		return nil
	}

	// Non-open method on an INIT-state channel is an error
	if channel.state == CH_STATE_INIT && (classId != 20 || methodId != 10) {
		return amqp.NewHardError(
			503,
			"Non-Channel.Open method called on unopened channel",
			classId,
			methodId,
		)
	}
	// Route
	// fmt.Println("Routing method: " + methodFrame.MethodName())
	switch {
	case classId == 10:
		return channel.connectionRoute(channel.conn, methodFrame)
	case classId == 20:
		return channel.channelRoute(methodFrame)
	case classId == 40:
		return channel.exchangeRoute(methodFrame)
	case classId == 50:
		return channel.queueRoute(methodFrame)
	case classId == 60:
		return channel.basicRoute(methodFrame)
	case classId == 90:
		return channel.txRoute(methodFrame)
	case classId == 85:
		return channel.confirmRoute(methodFrame)
	default:
		return amqp.NewHardError(540, "Not implemented", classId, methodId)
	}
	return nil
}
