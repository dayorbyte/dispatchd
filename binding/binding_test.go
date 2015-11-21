package binding

import (
	"github.com/jeffjenkins/mq/amqp"
	"testing"
)

func TestEquals(t *testing.T) {
	var bNil *Binding = nil
	var b = NewBinding("q1", "e1", "rk", amqp.NewTable())
	var same = NewBinding("q1", "e1", "rk", amqp.NewTable())
	var diffQ = NewBinding("DIFF", "e1", "rk", amqp.NewTable())
	var diffE = NewBinding("q1", "DIFF", "rk", amqp.NewTable())
	var diffR = NewBinding("q1", "e1", "DIFF", amqp.NewTable())
	if b.Equals(nil) || bNil.Equals(b) {
		t.Errorf("Comparison to nil was true!")
	}
	if !b.Equals(same) {
		t.Errorf("Equals returns false!")
	}
	if b.Equals(diffQ) {
		t.Errorf("Equals returns true on queue name diff!")
	}
	if b.Equals(diffE) {
		t.Errorf("Equals returns true on exchange name diff!")
	}
	if b.Equals(diffR) {
		t.Errorf("Equals returns true on routing key diff!")
	}

}
