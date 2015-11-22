package amqp

import (
	"bytes"
	"reflect"
	"testing"
)

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
	var outBind = &ExchangeBind{
		Destination: "dest",
		Source:      "src",
		RoutingKey:  "rk",
		NoWait:      true,
		Arguments:   NewTable(), //everythingTable(),
	}
	var outBuf = bytes.NewBuffer([]byte{})
	err := outBind.Write(outBuf)
	if err != nil {
		t.Errorf(err.Error())
	}
	var outBytes = outBuf.Bytes()[4:]
	printWireBytes(outBytes, t)
	var inBind = &ExchangeBind{}
	err = inBind.Read(bytes.NewBuffer(outBytes))
	if err != nil {
		t.Errorf(err.Error())
	}
}

func printWireBytes(bs []byte, t *testing.T) {
	t.Logf("Byte count: %d", len(bs))
	for _, b := range bs {
		t.Logf("(%c %d),", b, b)
	}
	t.Logf("\n")
}
