package binding

import (
	"github.com/jeffjenkins/mq/amqp"
	"testing"
)

func TestEquals(t *testing.T) {
	var bNil *Binding = nil
	b, _ := NewBinding("q1", "e1", "rk", amqp.NewTable(), false)
	same, _ := NewBinding("q1", "e1", "rk", amqp.NewTable(), false)
	diffQ, _ := NewBinding("DIFF", "e1", "rk", amqp.NewTable(), false)
	diffE, _ := NewBinding("q1", "DIFF", "rk", amqp.NewTable(), false)
	diffR, _ := NewBinding("q1", "e1", "DIFF", amqp.NewTable(), false)

	if b == nil || same == nil || diffQ == nil || diffE == nil || diffR == nil {
		t.Errorf("Failed to construct bindings")
	}
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

func TestTopicRouting(t *testing.T) {
	_, err := NewBinding("q1", "e1", "(", amqp.NewTable(), true)
	if err == nil {
		t.Errorf("Bad topic patter compiled!")
	}
	basic, _ := NewBinding("q1", "e1", "hello.world", amqp.NewTable(), true)
	singleWild, _ := NewBinding("q1", "e1", "hello.*.world", amqp.NewTable(), true)
	multiWild, _ := NewBinding("q1", "e1", "hello.#.world", amqp.NewTable(), true)
	multiWild2, _ := NewBinding("q1", "e1", "hello.#.world.#", amqp.NewTable(), true)
	if !basic.MatchTopic(basicPublish("e1", "hello.world")) {
		t.Errorf("Basic match failed")
	}
	if basic.MatchTopic(basicPublish("e1", "hello.worlds")) {
		t.Errorf("Incorrect match with suffix")
	}
	if !basic.MatchTopic(basicPublish("e1", "hello.world")) {
		t.Errorf("Match succeeded despite mismatched exchange")
	}
	if !singleWild.MatchTopic(basicPublish("e1", "hello.one.world")) {
		t.Errorf("Failed to match single wildcard")
	}
	if singleWild.MatchTopic(basicPublish("e1", "hello.world")) {
		t.Errorf("Matched without wildcard token")
	}
	if !multiWild.MatchTopic(basicPublish("e1", "hello.one.two.three.world")) {
		t.Errorf("Failed to match multi wildcard")
	}
	if !multiWild2.MatchTopic(basicPublish("e1", "hello.one.world.hi")) {
		t.Errorf("Multiple multi-wild tokens failed")
	}

}

func basicPublish(e string, key string) *amqp.BasicPublish {
	return &amqp.BasicPublish{
		Exchange:   e,
		RoutingKey: key,
	}
}
