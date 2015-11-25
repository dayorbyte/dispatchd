package amqp

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

var testRand *rand.Rand = nil

func init() {
	var source = rand.NewSource(int64(1234))
	testRand = rand.New(source)
}

func TestMalformedTable(t *testing.T) {
	var table = NewTable()
	table.SetKey("hi", "bye")
	table.SetKey("mi", "bye")
	var method = &ExchangeBind{
		Destination: "dest",
		Source:      "src",
		RoutingKey:  "rk",
		NoWait:      true,
		Arguments:   table,
	}
	var outBuf = bytes.NewBuffer([]byte{})
	err := method.Write(outBuf)
	if err != nil {
		t.Errorf(err.Error())
	}
	outBytes := outBuf.Bytes()[4:]
	// Use this to see which bytes to alter
	printWireBytes(outBytes, t)

	// make key too long
	outBytes[19] = 250
	var inMethod = &ExchangeBind{}
	err = inMethod.Read(bytes.NewBuffer(outBytes))
	if err == nil {
		t.Errorf("Successfully read malformed bytes!")
	}
	outBytes[19] = 2

	// make value the wrong type
	outBytes[22] = 'S'
	inMethod = &ExchangeBind{}
	err = inMethod.Read(bytes.NewBuffer(outBytes))
	if err == nil {
		t.Errorf("Successfully read malformed bytes!")
	}
	outBytes[22] = 's'

	// duplicate keys
	outBytes[28] = 'h'
	inMethod = &ExchangeBind{}
	err = inMethod.Read(bytes.NewBuffer(outBytes))
	if err == nil {
		t.Errorf("Successfully read malformed bytes!")
	}
	outBytes[28] = 'm'

	// can't read value type
	inMethod = &ExchangeBind{}
	outBytes[27] = 7
	err = inMethod.Read(bytes.NewBuffer(outBytes))
	if err == nil {
		t.Errorf("Successfully read malformed bytes!")
	}
	outBytes[27] = 2

	// t.Errorf("fail")
}

func TestReadValue(t *testing.T) {
	tryRead := func(bts []byte, msg string) {
		_, err := readValue(bytes.NewBuffer(bts))
		if err == nil {
			t.Errorf(msg)
		}
	}
	var types = []byte{'t', 'b', 'B', 'U', 'u', 'I', 'i', 'L', 'l',
		'f', 'd', 'D', 's', 'S', 'A', 'T', 'F'}
	// 'V' isn't included since it has no value
	for i := 0; i < len(types); i++ {
		tryRead([]byte{types[i]}, fmt.Sprintf("Successfully read malformed value for type %c", types[i]))
	}
	tryRead([]byte{'D', 1}, "read malformed decimal")

	tryRead([]byte{'~', 1}, "read bad value type")

	// successful timestamp read since the server doesn't really support them
	val, err := readValue(bytes.NewBuffer([]byte{'T', 0, 0, 0, 0, 0, 0, 0, 2}))
	if err != nil {
		t.Errorf(err.Error())
	}
	if val.GetVTimestamp() != uint64(2) {
		t.Errorf("Failed to deserialize uint64")
	}
	// successful 'V' (no value) read, mainly for coverage
	val, err = readValue(bytes.NewBuffer([]byte{'V'}))
	if err != nil {
		t.Errorf(err.Error())
	}
	if val != nil {
		t.Errorf("Failed to deserialize 'V' field ")
	}
	// tryRead()

}

func sptr(s string) *string {
	return &s
}

func bptr(b byte) *byte {
	return &b
}

// func allFlags() uint16 {
// 	return (MaskContentType | MaskContentEncoding | MaskHeaders |
// 		MaskDeliveryMode | MaskPriority | MaskCorrelationId | MaskReplyTo |
// 		MaskExpiration | MaskMessageId | MaskTimestamp | MaskType | MaskUserId |
// 		MaskAppId | MaskReserved)
// }

func TestReadingContentHeaderProps(t *testing.T) {
	time := uint64(1312312)
	var props = BasicContentHeaderProperties{
		ContentType:     sptr("ContentType"),
		ContentEncoding: sptr("ContentEncoding"),
		Headers:         NewTable(),
		DeliveryMode:    bptr(byte(1)),
		Priority:        bptr(byte(1)),
		CorrelationId:   sptr("CorrelationId"),
		ReplyTo:         sptr("ReplyTo"),
		Expiration:      sptr("Expiration"),
		MessageId:       sptr("MessageId"),
		Timestamp:       &time,
		Type:            sptr("Type"),
		UserId:          sptr("UserId"),
		AppId:           sptr("AppId"),
		Reserved:        sptr(""),
	}
	var outBuf = bytes.NewBuffer([]byte{})
	flags, err := props.WriteProps(outBuf)
	if err != nil {
		t.Errorf(err.Error())
	}
	outBytes := outBuf.Bytes()
	// Use subsets of bytes to trigger all failure conditions
	for i := 0; i < len(outBytes); i++ {
		var partialBuffer = bytes.NewBuffer(outBytes[:i])
		var inProps = &BasicContentHeaderProperties{}
		err = inProps.ReadProps(flags, partialBuffer)
		if err == nil {
			t.Errorf("Successfully read malformed props. %d/%d bytes read", i, len(outBytes))
		}
	}
	// Succeed in reading all bytes
	var partialBuffer = bytes.NewBuffer(outBytes)
	var inProps = &BasicContentHeaderProperties{}
	err = inProps.ReadProps(flags, partialBuffer)
	if err != nil {
		t.Errorf(err.Error())
	}
	// check a few random fields
	if *inProps.ContentType != "ContentType" {
		t.Errorf("Bad content type: %s", *inProps.ContentType)
	}
	if *inProps.Timestamp != time {
		t.Errorf("Bad timestamp: %s", *inProps.Timestamp)
	}
}

func TestReadArrayFailures(t *testing.T) {
	_, err := readArray(bytes.NewBuffer([]byte{0, 0, 0, 1, 'L'}))
	if err == nil {
		t.Errorf("Read a malformed array value")
	}
}

func TestWireFrame(t *testing.T) {
	// Write frame to bytes
	var outFrame = &WireFrame{
		FrameType: uint8(10),
		Channel:   uint16(12311),
		Payload:   []byte{0, 0, 9, 1},
	}
	var buf = bytes.NewBuffer(make([]byte, 0))
	WriteFrame(buf, outFrame)

	// Read frame from bytes
	var outBytes = buf.Bytes()
	var inFrame, err = ReadFrame(bytes.NewBuffer(outBytes))
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(inFrame, outFrame) {
		t.Errorf("Couldn't read the frame that was written")
	}
	// Incomplete frames
	for i := 0; i < len(outBytes); i++ {
		var noTypeBuf = bytes.NewBuffer(outBytes[:i])
		_, err = ReadFrame(noTypeBuf)
		if err == nil {
			t.Errorf("No error on malformed frame. %d/%d bytes read", i, len(buf.Bytes()))
		}
	}
}

func TestMethodTypes(t *testing.T) {
	for _, method := range methodsForTesting() {
		var outBuf = bytes.NewBuffer([]byte{})
		err := method.Write(outBuf)
		if err != nil {
			t.Errorf(err.Error())
		}

		var outBytes = outBuf.Bytes()[4:]
		// Try all lengths of bytes below the ones needed
		for index, _ := range outBytes {
			var inBind = reflect.New(reflect.TypeOf(method).Elem()).Interface().(MethodFrame)
			err = inBind.Read(bytes.NewBuffer(outBytes[:index]))
			if err == nil {
				printWireBytes(outBytes[:index], t)
				t.Errorf("Parsed malformed request bytes")
				return
			}
		}
		// printWireBytes(outBytes, t)
		// Try the right set of bytes
		var inBind = reflect.New(reflect.TypeOf(method).Elem()).Interface().(MethodFrame)
		err = inBind.Read(bytes.NewBuffer(outBytes))
		if err != nil {
			t.Logf("Method is %s", method.MethodName())
			printWireBytes(outBytes, t)
			t.Errorf(err.Error())
			return
		}
	}
}

func methodsForTesting() []MethodFrame {
	return []MethodFrame{
		&ExchangeBind{
			Destination: "dest",
			Source:      "src",
			RoutingKey:  "rk",
			NoWait:      true,
			Arguments:   EverythingTable(),
		},
		&ConnectionTune{
			ChannelMax: uint16(3),
			FrameMax:   uint32(1),
			Heartbeat:  uint16(2),
		},
		&BasicDeliver{
			ConsumerTag: string("deliver"),
			DeliveryTag: uint64(4),
			Redelivered: true,
			Exchange:    string("ex1"),
			RoutingKey:  string("rk1"),
		},
	}
}

func printWireBytes(bs []byte, t *testing.T) {
	t.Logf("Byte count: %d", len(bs))
	for i, b := range bs {
		t.Logf("%d:(%c %d),", i, b, b)
	}
	t.Logf("\n")
}
