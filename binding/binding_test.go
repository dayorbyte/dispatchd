package binding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
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
	var bind = &amqp.QueueBind{
		Exchange:   "ex1",
		Queue:      "q1",
		RoutingKey: "rk1",
		Arguments:  amqp.NewTable(),
	}
	PersistBinding(db, bind)

	// Read
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("bindings"))
		if bucket == nil {
			return fmt.Errorf("No bucket!")
		}
		// iterate through queues
		cursor := bucket.Cursor()
		var count = 0
		for name, data := cursor.First(); name != nil; name, data = cursor.Next() {
			count += 1
			// Read
			var method = &amqp.QueueBind{}
			var reader = bytes.NewReader(data)
			var err = method.Read(reader)
			if err != nil {
				return err
			}
			if !(method.Queue == bind.Queue && method.Exchange == bind.Exchange && method.RoutingKey == bind.RoutingKey) {
				return fmt.Errorf("Loaded binding did not match")
			}
		}
		if count != 1 {
			return fmt.Errorf("Too many database bindings. Expected 1")
		}
		if count == 0 {
			return fmt.Errorf("Too few database bindings. Expected 1")
		}
		return nil
	})
	if err != nil {
		t.Errorf(err.Error())
	}

	// Depersist
	b, err := NewBinding(bind.Queue, bind.Exchange, bind.RoutingKey, amqp.NewTable(), false)
	if err != nil {
		t.Errorf("Error constructing binding")
	}
	b.Depersist(db)

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("bindings"))
		if bucket == nil {
			return fmt.Errorf("No bucket!")
		}
		// iterate through queues
		cursor := bucket.Cursor()
		var count = 0
		for name, _ := cursor.First(); name != nil; name, _ = cursor.Next() {
			count += 1
		}
		if count != 0 {
			return fmt.Errorf("Too many database bindings. Expected 0")
		}
		return nil
	})
	if err != nil {
		t.Errorf(err.Error())
	}

}
