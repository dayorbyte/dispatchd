package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jeffjenkins/mq/amqp"
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
	currentMessage *Message
	consumers      map[string]*Consumer
	consumerLock   sync.Mutex
	sendLock       sync.Mutex
	lastQueueName  string
	flow           bool
	txMode         bool
	txLock         sync.Mutex
	txMessages     []*TxMessage
	txAcks         []*TxAck
	// Consumers
	msgIndex uint64
	// Delivery Tracking
	deliveryTag  uint64
	deliveryLock sync.Mutex
	ackLock      sync.Mutex
	awaitingAcks map[uint64]UnackedMessage
	// Channel QOS Limits
	limitLock     sync.Mutex
	prefetchSize  uint32
	prefetchCount uint16
	activeSize    uint32
	activeCount   uint16
	// Consumer default QOS limits
	defaultPrefetchSize  uint32
	defaultPrefetchCount uint16
}

func (channel *Channel) commitTx() {
	channel.txLock.Lock()
	defer channel.txLock.Unlock()
	// messages
	// todo: persist all messages to all queues once persistence exists
	for _, msg := range channel.txMessages {
		queue, found := channel.server.queues[msg.queueName]
		// the
		if !found {
			continue
		}
		queue.add(msg.msg)
	}
	// Acks
	// todo: remove acked messages from persistent storage in a single
	// transaction
	for _, ack := range channel.txAcks {
		channel.txAckMessage(ack)
	}

	// Clear transaction
	channel.txMessages = make([]*TxMessage, 0)
	channel.txAcks = make([]*TxAck, 0)
}

func (channel *Channel) txAckMessage(ack *TxAck) {
	switch {
	case ack.multiple && !ack.nack:
		channel.ackBelow(ack.id, true)
	case ack.multiple && ack.nack:
		channel.nackBelow(ack.id, ack.requeueNack, true)
	case !ack.multiple && !ack.nack:
		channel.ackOne(ack.id, true)
	case !ack.multiple && ack.nack:
		channel.nackOne(ack.id, ack.requeueNack, true)
	}
}

func (channel *Channel) rollbackTx() {
	channel.txLock.Lock()
	defer channel.txLock.Unlock()
	channel.txMessages = make([]*TxMessage, 0)
	channel.txAcks = make([]*TxAck, 0)
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
			unacked.msg.redelivered += 1
			unacked.queue.readd(unacked.msg)
			var size = unacked.msg.size()
			channel.decrActive(1, size)
			unacked.consumer.decrActive(1, size)
		}
		// Clear awaiting acks
		channel.awaitingAcks = make(map[uint64]UnackedMessage)
	} else {
		// Redeliver. Don't need to mess with stats.
		// We do this in a short-lived goroutine since this could end up
		// blocking on sending to the network inside the consumer
		go func() {
			for tag, unacked := range channel.awaitingAcks {
				unacked.msg.redelivered += 1
				unacked.consumer.redeliver(tag, unacked.msg)
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
			consumer.ping()
		}
	}
}

func (channel *Channel) channelErrorWithMethod(code uint16, message string, classId uint16, methodId uint16) {
	fmt.Println("Sending channel error:", message)
	channel.state = CH_STATE_CLOSING
	channel.sendMethod(&amqp.ChannelClose{code, message, classId, methodId})
}

func (channel *Channel) ackBelow(tag uint64, commitTx bool) bool {
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, &TxAck{
			id:          tag,
			nack:        false,
			requeueNack: false,
			multiple:    true,
		})
		return true
	}
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	var count = 0
	for k, unacked := range channel.awaitingAcks {
		if k <= tag || tag == 0 {
			delete(channel.awaitingAcks, k)
			var size = unacked.msg.size()
			unacked.consumer.decrActive(1, size)
			channel.decrActive(1, size)
			// TODO: select?
			unacked.consumer.ping()
			count += 1
		}
	}
	// TODO: should this be false if nothing was actually deleted and tag != 0?
	return true
}

func (channel *Channel) ackOne(tag uint64, commitTx bool) bool {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()

	var unacked, found = channel.awaitingAcks[tag]
	if !found {
		return false
	}
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, &TxAck{
			id:          tag,
			nack:        false,
			requeueNack: false,
			multiple:    false,
		})
		return true
	}
	delete(channel.awaitingAcks, tag)
	var size = unacked.msg.size()
	unacked.consumer.decrActive(1, size)
	channel.decrActive(1, size)
	unacked.msg.size()
	unacked.consumer.ping()
	return true
}

func (channel *Channel) nackBelow(tag uint64, requeue bool, commitTx bool) bool {
	// fmt.Println("Nack below")
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, &TxAck{
			id:          tag,
			nack:        true,
			requeueNack: requeue,
			multiple:    true,
		})
		return true
	}
	var count = 0
	for k, unacked := range channel.awaitingAcks {
		fmt.Printf("%d(%d), ", k, tag)
		if k <= tag || tag == 0 {
			delete(channel.awaitingAcks, k)
			if requeue {
				unacked.consumer.queue.readd(unacked.msg)
			}
			var size = unacked.msg.size()
			unacked.consumer.decrActive(1, size)
			channel.decrActive(1, size)
			// TODO: select?
			unacked.consumer.ping()
			count += 1
		}
	}
	// fmt.Println()
	// fmt.Printf("Nacked %d messages\n", count)
	// TODO: should this be false if nothing was actually deleted and tag != 0?
	return true
}

func (channel *Channel) nackOne(tag uint64, requeue bool, commitTx bool) bool {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	var unacked, found = channel.awaitingAcks[tag]
	if !found {
		return false
	}
	if channel.txMode && !commitTx {
		channel.txLock.Lock()
		defer channel.txLock.Unlock()
		channel.txAcks = append(channel.txAcks, &TxAck{
			id:          tag,
			nack:        true,
			requeueNack: requeue,
			multiple:    false,
		})
		return true
	}
	if requeue {
		unacked.consumer.queue.readd(unacked.msg)
	}
	var size = unacked.msg.size()
	unacked.consumer.decrActive(1, size)
	channel.decrActive(1, size)
	// TODO: select?
	unacked.consumer.ping()
	delete(channel.awaitingAcks, tag)
	return true
}

func (channel *Channel) addUnackedMessage(consumer *Consumer, msg *Message) uint64 {
	var tag = channel.nextDeliveryTag()
	var unacked = UnackedMessage{
		consumer: consumer,
		msg:      msg,
	}
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()

	_, found := channel.awaitingAcks[tag]
	if found {
		panic(fmt.Sprintf("Already found tag: %s", tag))
	}
	channel.awaitingAcks[tag] = unacked
	// fmt.Printf("Adding tag: %d\n", tag)
	return tag
}

func (channel *Channel) addConsumer(consumer *Consumer) error {
	channel.consumerLock.Lock()
	defer channel.consumerLock.Unlock()
	_, found := channel.consumers[consumer.consumerTag]
	if found {
		return fmt.Errorf("Consumer tag already exists: %s", consumer.consumerTag)
	}
	channel.consumers[consumer.consumerTag] = consumer
	return nil
}

func (channel *Channel) decrActive(count uint16, size uint32) {
	channel.limitLock.Lock()
	channel.activeCount -= count
	channel.activeSize -= size
	channel.limitLock.Unlock()
}

func (channel *Channel) acquireResources(count uint16, size uint32) bool {
	channel.limitLock.Lock()
	defer channel.limitLock.Unlock()
	var sizeOk = channel.prefetchSize == 0 || channel.activeSize < channel.prefetchSize
	var countOk = channel.prefetchCount == 0 || channel.activeCount < channel.prefetchCount
	// If we're OK on size and count, acquire the resources
	if sizeOk && countOk {
		channel.activeCount += count
		channel.activeSize += size
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
	channel.currentMessage = NewMessage(method, channel.conn.id)
	return nil
}

func NewChannel(id uint16, conn *AMQPConnection) *Channel {
	return &Channel{
		id:           id,
		server:       conn.server,
		incoming:     make(chan *amqp.WireFrame),
		outgoing:     conn.outgoing,
		conn:         conn,
		flow:         true,
		state:        CH_STATE_INIT,
		txMessages:   make([]*TxMessage, 0),
		txAcks:       make([]*TxAck, 0),
		consumers:    make(map[string]*Consumer),
		awaitingAcks: make(map[uint64]UnackedMessage),
	}
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
			switch {
			case frame.FrameType == uint8(amqp.FrameMethod):
				channel.routeMethod(frame)
			case frame.FrameType == uint8(amqp.FrameHeader):
				if channel.state != CH_STATE_CLOSING {
					channel.handleContentHeader(frame)
				}
			case frame.FrameType == uint8(amqp.FrameBody):
				if channel.state != CH_STATE_CLOSING {
					channel.handleContentBody(frame)
				}
			default:
				fmt.Println("Unknown frame type!")
			}
		}
	}()
}

func (channel *Channel) startChannel() {

}

func (channel *Channel) close(code uint16, reason string, classId uint16, methodId uint16) {
	channel.sendMethod(&amqp.ChannelClose{
		ReplyCode: code,
		ReplyText: reason,
		ClassId:   classId,
		MethodId:  methodId,
	})
	channel.state = CH_STATE_CLOSING
}

func (channel *Channel) shutdown() {
	fmt.Printf("Shutdown called on channel %d\n", channel.id)
	if channel.state == CH_STATE_CLOSED {
		fmt.Printf("Shutdown already finished on %d\n", channel.id)
		return
	}
	channel.state = CH_STATE_CLOSED
	// unregister this channel
	channel.conn.deregisterChannel(channel.id)
	// remove any consumers associated with this channel
	fmt.Printf("Stop channel consumers %d\n", channel.id)
	for _, consumer := range channel.consumers {
		consumer.stop()
	}
	// Any unacked messages should be re-added
	for _, unacked := range channel.awaitingAcks {
		// TODO(MUST): If we want at-most-once delivery we can't re-add these
		// messages. Need to figure out if the spec specifies, and after that
		// provide a way to have both options. Maybe a message header?
		unacked.consumer.queue.readd(unacked.msg)
		// this probably isn't needed, but for debugging purposes it's nice to
		// ensure that all the active counts/sizes get back to 0
		var size = unacked.msg.size()
		unacked.consumer.decrActive(1, size)
		channel.decrActive(1, size)
	}
}

func (channel *Channel) removeConsumer(consumerTag string) error {
	// TODO: how does this interact with the code in shutdown?
	var consumer, found = channel.consumers[consumerTag]
	if !found {
		return errors.New("Consumer not found")
	}
	consumer.stop()
	delete(channel.consumers, consumerTag)
	return nil
}

// Send a method frame out to the client
// TODO: why isn't this taking a pointer?
func (channel *Channel) sendMethod(method amqp.MethodFrame) {
	// fmt.Printf("Sending method: %s\n", method.MethodName())
	var buf = bytes.NewBuffer([]byte{})
	method.Write(buf)
	channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameMethod), channel.id, buf.Bytes()}
}

// Send a method frame out to the client
func (channel *Channel) sendContent(method amqp.MethodFrame, message *Message) {
	channel.sendLock.Lock()
	defer channel.sendLock.Unlock()
	// deliver
	channel.sendMethod(method)
	// header
	channel.outgoing <- &amqp.WireFrame{uint8(amqp.FrameHeader), channel.id, message.header.AsBytes}
	// body
	for _, b := range message.payload {
		b.Channel = channel.id
		channel.outgoing <- b
	}
}

func (channel *Channel) handleContentHeader(frame *amqp.WireFrame) {
	if channel.currentMessage == nil {
		// TODO: error
		fmt.Println("Unexpected content header frame!")
		return
	}
	if channel.currentMessage.header != nil {
		// TODO: error
		fmt.Println("Unexpected content header frame! Already saw header")
	}
	var headerFrame = &amqp.ContentHeaderFrame{}
	var err = headerFrame.Read(bytes.NewReader(frame.Payload))
	headerFrame.AsBytes = frame.Payload
	if err != nil {
		// TODO: error
		fmt.Println("Error parsing header frame: " + err.Error())
	}

	channel.currentMessage.header = headerFrame
}

func (channel *Channel) handleContentBody(frame *amqp.WireFrame) {
	if channel.currentMessage == nil {
		// TODO: error
		fmt.Println("Unexpected content body frame. No method content-having method called yet!")
	}
	if channel.currentMessage.header == nil {
		// TODO: error
		fmt.Println("Unexpected content body frame! No header yet")
	}
	channel.currentMessage.payload = append(channel.currentMessage.payload, frame)
	// TODO: store this on message
	var size = uint64(0)
	for _, body := range channel.currentMessage.payload {
		size += uint64(len(body.Payload))
	}
	if size < channel.currentMessage.header.ContentBodySize {
		return
	}

	// We have the whole contents, let's publish!
	var server = channel.server
	var message = channel.currentMessage

	exchange, _ := server.exchanges[message.method.Exchange]

	if channel.txMode {
		// TxMode, add the messages to a list
		queues := exchange.queuesForPublish(server, channel, channel.currentMessage)
		channel.txLock.Lock()
		for queueName, _ := range queues {
			var txmsg = &TxMessage{
				msg:       message,
				queueName: queueName,
			}
			channel.txMessages = append(channel.txMessages, txmsg)
		}
		channel.txLock.Unlock()
	} else {
		// Normal mode, publish directly
		exchange.publish(server, channel, channel.currentMessage)
	}

	channel.currentMessage = nil
	if channel.confirmMode {
		channel.msgIndex += 1
		channel.sendMethod(&amqp.BasicAck{channel.msgIndex, false})
	}
}

func (channel *Channel) routeMethod(frame *amqp.WireFrame) error {
	var methodReader = bytes.NewReader(frame.Payload)
	var methodFrame, err = amqp.ReadMethod(methodReader)
	if err != nil {
		fmt.Println("ERROR: ", err)
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
		channel.conn.connectionErrorWithMethod(
			503,
			"Non-Channel.Open method called on unopened channel",
			classId,
			methodId,
		)
		return nil
	}
	// Route
	// fmt.Println("Routing method: " + methodFrame.MethodName())
	switch {
	case classId == 10:
		channel.connectionRoute(channel.conn, methodFrame)
	case classId == 20:
		channel.channelRoute(methodFrame)
	case classId == 40:
		channel.exchangeRoute(methodFrame)
	case classId == 50:
		channel.queueRoute(methodFrame)
	case classId == 60:
		channel.basicRoute(methodFrame)
	case classId == 90:
		channel.txRoute(methodFrame)
	case classId == 85:
		channel.confirmRoute(methodFrame)
	default:
		channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	}
	return nil
}
