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

func (channel *Channel) channelError(code uint16, message string) {
	channel.channelErrorWithMethod(code, message, 0, 0)
}

func (channel *Channel) channelErrorWithMethod(code uint16, message string, classId uint16, methodId uint16) {
	fmt.Println("Sending channel error:", message)
	channel.state = CH_STATE_CLOSING
	channel.sendMethod(&amqp.ChannelClose{code, message, classId, methodId})
}

func (channel *Channel) ackBelow(tag uint64) bool {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	fmt.Println("Ack below")
	var count = 0
	for k, unacked := range channel.awaitingAcks {
		// fmt.Printf("%d(%d), ", k, tag)
		if k <= tag || tag == 0 {
			delete(channel.awaitingAcks, k)
			unacked.consumer.decrActive(1, unacked.msg.size())
			count += 1
		}
	}
	// fmt.Println()
	// fmt.Printf("Acked %d messages\n", count)
	// TODO: should this be false if nothing was actually deleted and tag != 0?
	return true
}

func (channel *Channel) ackOne(tag uint64) bool {
	// fmt.Println("Ack one")
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	var unacked, found = channel.awaitingAcks[tag]
	if !found {
		return false
	}
	delete(channel.awaitingAcks, tag)
	unacked.consumer.decrActive(1, unacked.msg.size())
	return true
}

func (channel *Channel) nackBelow(tag uint64) bool {
	// fmt.Println("Nack below")
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	var count = 0
	for k, unacked := range channel.awaitingAcks {
		fmt.Printf("%d(%d), ", k, tag)
		if k <= tag || tag == 0 {
			delete(channel.awaitingAcks, k)
			unacked.consumer.queue.readd(unacked.msg)
			unacked.consumer.decrActive(1, unacked.msg.size())
			count += 1
		}
	}
	// fmt.Println()
	// fmt.Printf("Nacked %d messages\n", count)
	// TODO: should this be false if nothing was actually deleted and tag != 0?
	return true
}

func (channel *Channel) nackOne(tag uint64) bool {
	channel.ackLock.Lock()
	defer channel.ackLock.Unlock()
	var unacked, found = channel.awaitingAcks[tag]
	if !found {
		return false
	}
	unacked.consumer.queue.readd(unacked.msg)
	unacked.consumer.decrActive(1, unacked.msg.size())
	delete(channel.awaitingAcks, tag)
	return true
}

func (channel *Channel) addUnackedMessage(consumer *Consumer, msg *Message) uint64 {
	// fmt.Println("Adding unacked message")
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

func (channel *Channel) addConsumer(consumer *Consumer) {
	// TODO: error handling
	channel.consumers[consumer.consumerTag] = consumer
}

func (channel *Channel) consumeLimitsOk() bool {
	var sizeOk = channel.prefetchSize == 0 || channel.activeSize <= channel.prefetchSize
	var countOk = channel.prefetchCount == 0 || channel.activeCount <= channel.prefetchCount
	// fmt.Printf("%d|%d || %d|%d\n", channel.prefetchSize, channel.activeSize, channel.prefetchCount, channel.activeCount)
	return sizeOk && countOk
}

func (channel *Channel) incrActive(count uint16, size uint32) {
	channel.limitLock.Lock()
	channel.activeCount += count
	channel.activeSize += size
	channel.limitLock.Unlock()
}

func (channel *Channel) decrActive(count uint16, size uint32) {
	channel.limitLock.Lock()
	channel.activeCount -= count
	channel.activeSize -= size
	channel.limitLock.Unlock()
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
	channel.currentMessage = NewMessage(method)
	return nil
}

func NewChannel(id uint16, conn *AMQPConnection) *Channel {
	return &Channel{
		id:           id,
		server:       conn.server,
		incoming:     make(chan *amqp.WireFrame),
		outgoing:     conn.outgoing,
		conn:         conn,
		state:        CH_STATE_INIT,
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
	channel.deliveryLock.Unlock()
	return channel.deliveryTag
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
				channel.handleContentHeader(frame)
			case frame.FrameType == uint8(amqp.FrameBody):
				channel.handleContentBody(frame)
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
		unacked.consumer.queue.readd(unacked.msg)
		// this probably isn't needed, but for debugging purposes it's nice to
		// ensure that all the active counts/sizes get back to 0
		unacked.consumer.decrActive(1, unacked.msg.size())
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
	// fmt.Println("Sending content\n")
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
	var server = channel.server
	var message = channel.currentMessage
	server.exchanges[message.method.Exchange].publish(server, channel, channel.currentMessage)
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
	case classId == 85:
		channel.confirmRoute(methodFrame)
	default:
		channel.conn.connectionErrorWithMethod(540, "Not implemented", classId, methodId)
	}
	return nil
}
