package binding

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/dispatchd/amqp"
	// "github.com/jeffjenkins/dispatchd/persist"
	"os"
	"testing"
)

func TestFanout(t *testing.T) {
	b, _ := NewBinding("q1", "e1", "rk", amqp.NewTable(), false)
	if b.MatchFanout(basicPublish("DIFF", "asdf")) {
		t.Errorf("Fanout did not check exchanges")
	}
	if !b.MatchFanout(basicPublish("e1", "asdf")) {
		t.Errorf("Fanout didn't match regardless of key")
	}
}

func TestDirect(t *testing.T) {
	b, _ := NewBinding("q1", "e1", "rk", amqp.NewTable(), false)
	if b.MatchDirect(basicPublish("DIFF", "asdf")) {
		t.Errorf("MatchDirect did not check exchanges")
	}
	if b.MatchDirect(basicPublish("e1", "asdf")) {
		t.Errorf("MatchDirect matched even with the wrong key")
	}
	if !b.MatchDirect(basicPublish("e1", "rk")) {
		t.Errorf("MatchDirect did not match with the correct key and exchange")
	}
}

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

func TestJson(t *testing.T) {
	basic, _ := NewBinding("q1", "e1", "hello.world", amqp.NewTable(), true)
	var basicBytes, err = basic.MarshalJSON()
	if err != nil {
		t.Errorf(err.Error())
	}

	expectedBytes, err := json.Marshal(map[string]interface{}{
		"queueName":    "q1",
		"exchangeName": "e1",
		"key":          "hello.world",
		"arguments":    make(map[string]interface{}),
	})
	if string(expectedBytes) != string(basicBytes) {
		t.Logf("Expected: %s", expectedBytes)
		t.Logf("Got: %s", basicBytes)
		t.Errorf("Wrong json bytes!")
	}

}

func TestPersistence(t *testing.T) {
	var dbFile = "TestBindingPersistence.db"
	os.Remove(dbFile)
	defer os.Remove(dbFile)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		t.Errorf("Failed to create db")
	}

	// Persist
	b, err := NewBinding("q1", "ex1", "rk1", amqp.NewTable(), true)
	if err != nil {
		t.Errorf("Error in NewBinding")
	}
	err = b.Persist(db)
	if err != nil {
		t.Errorf("Error in NewBinding")
	}

	// Read
	bMap, err := LoadAllBindings(db)
	if err != nil {
		t.Errorf("Error in LoadAllBindings")
	}
	if len(bMap) != 1 {
		t.Errorf("Wrong number of bindings")
	}
	for _, b2 := range bMap {
		if !b2.Equals(b) {
			t.Errorf("Did not get the same binding from the db")
		}
	}

	// Depersist
	b.Depersist(db)

	bMap, err = LoadAllBindings(db)
	if err != nil {
		t.Errorf("Error in LoadAllBindings")
	}
	if len(bMap) != 0 {
		t.Errorf("Wrong number of bindings")
	}

}
