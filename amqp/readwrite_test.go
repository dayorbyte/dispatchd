package amqp

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
)

var testRand *rand.Rand = nil

func init() {
	var source = rand.NewSource(int64(1234))
	testRand = rand.New(source)
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
			Arguments:   everythingTable(),
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
