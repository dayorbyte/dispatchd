package exchange

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/mq/amqp"
	"os"
	"reflect"
	"testing"
)

// EX_TYPE_DIRECT
// EX_TYPE_FANOUT
// EX_TYPE_TOPIC
// EX_TYPE_HEADERS

func TestClose(t *testing.T) {
	var ex = NewExchange(
		"ex",
		EX_TYPE_TOPIC,
		false,
		false,
		false,
		amqp.NewTable(),
		false,
		make(chan *Exchange),
	)
	if ex.Closed {
		t.Errorf("Exchange closed when it shouldn't be!")
	}
	ex.Close()
	if !ex.Closed {
		t.Errorf("Exchange closed when it shouldn't be!")
	}
}

func TestJSON(t *testing.T) {
	var ex = NewExchange(
		"ex",
		EX_TYPE_TOPIC,
		false,
		false,
		false,
		amqp.NewTable(),
		false,
		make(chan *Exchange),
	)
	var expected, err = json.Marshal(map[string]interface{}{
		"type":     "topic",
		"bindings": make([]int, 0),
	})
	if err != nil {
		t.Errorf(err.Error())
	}
	got, err := json.Marshal(ex)
	if err != nil {
		t.Errorf(err.Error())
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("unequal!\nexpected:%v\ngot     :%v", expected, got)
	}
	// ex.Extype = Extype(123)
}

func TestExchangeTypes(t *testing.T) {
	if ext, err := ExchangeNameToType("topic"); ext != EX_TYPE_TOPIC || err != nil {
		t.Errorf("Error converting type")
	}
	if ext, err := ExchangeNameToType("fanout"); ext != EX_TYPE_FANOUT || err != nil {
		t.Errorf("Error converting type")
	}
	if ext, err := ExchangeNameToType("direct"); ext != EX_TYPE_DIRECT || err != nil {
		t.Errorf("Error converting type")
	}
	if ext, err := ExchangeNameToType("headers"); ext != EX_TYPE_HEADERS || err != nil {
		t.Errorf("Error converting type")
	}
	if _, err := ExchangeNameToType("unknown!"); err == nil {
		t.Errorf("No error converting unknown exchange name")
	}
	//
	if ext, err := exchangeTypeToName(EX_TYPE_TOPIC); ext != "topic" || err != nil {
		t.Errorf("Error converting type")
	}
	if ext, err := exchangeTypeToName(EX_TYPE_FANOUT); ext != "fanout" || err != nil {
		t.Errorf("Error converting type")
	}
	if ext, err := exchangeTypeToName(EX_TYPE_DIRECT); ext != "direct" || err != nil {
		t.Errorf("Error converting type")
	}
	if ext, err := exchangeTypeToName(EX_TYPE_HEADERS); ext != "headers" || err != nil {
		t.Errorf("Error converting type")
	}
	if _, err := exchangeTypeToName(Extype(123)); err == nil {
		t.Errorf("No error converting bad type")
	}
}

func TestEquivalentExchanges(t *testing.T) {
	var ex = NewExchange(
		"ex1",
		EX_TYPE_DIRECT,
		true,
		true,
		true,
		amqp.NewTable(),
		true,
		make(chan *Exchange),
	)
	var ex2 = NewExchange(
		"ex1",
		EX_TYPE_DIRECT,
		true,
		true,
		true,
		amqp.NewTable(),
		true,
		make(chan *Exchange),
	)
	// Same
	if !ex.EquivalentExchanges(ex2) {
		t.Errorf("Same exchanges aren't equal!")
	}
	// name
	ex2.Name = "ex2"
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.Name = "ex1"

	// extype
	ex2.Extype = EX_TYPE_TOPIC
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.Extype = EX_TYPE_DIRECT
	// internal
	ex2.internal = false
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.internal = true
	// durable
	ex2.Durable = false
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.Durable = true
	// args
	var newTable = amqp.NewTable()
	newTable.SetKey("stuff", true)
	ex2.arguments = newTable
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.arguments = amqp.NewTable()
	// test other diffs ok
	ex2.System = false
	ex2.autodelete = false
	if !ex.EquivalentExchanges(ex2) {
		t.Errorf("Same exchanges aren't equal!")
	}
}

func TestExchangeRoutingDirect(t *testing.T) {
	// Make exchange ang binding
	var exDirect = NewExchange(
		"exd",
		EX_TYPE_DIRECT,
		false,
		false,
		false,
		amqp.NewTable(),
		false,
		make(chan *Exchange),
	)
	exDirect.AddBinding(&amqp.QueueBind{Queue: "q1", Exchange: "exd", RoutingKey: "rk-1"}, -1, false)

	// Create a random message, won't route by default
	var msg = amqp.RandomMessage(false)
	// Test wrong exchange for coverage
	res, err := exDirect.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	if len(res) > 0 {
		t.Errorf("Routed message which should not have routed", res)
	}

	// Test right exchange, wrong key
	msg.Method.Exchange = "exd"

	res, err = exDirect.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	if len(res) > 0 {
		t.Errorf("Routed message which should not have routed", res)
	}

	// Set the right values for routing
	msg.Method.RoutingKey = "rk-1"

	res, err = exDirect.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	if _, found := res["q1"]; !found {
		t.Errorf("Failed to route direct message: %s", res)
	}
}

func TestExchangeRoutingFanout(t *testing.T) {
	var exFanout = NewExchange(
		"exf",
		EX_TYPE_FANOUT,
		false,
		false,
		false,
		amqp.NewTable(),
		false,
		make(chan *Exchange),
	)
	exFanout.AddBinding(&amqp.QueueBind{Queue: "q1", Exchange: "exf", RoutingKey: "rk-1"}, -1, false)
	exFanout.AddBinding(&amqp.QueueBind{Queue: "q2", Exchange: "exf", RoutingKey: "rk-2"}, -1, false)

	// Create a random message, won't route by default
	var msg = amqp.RandomMessage(false)
	msg.Method.Exchange = "exf"

	res, err := exFanout.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	_, foundQ1 := res["q1"]
	_, foundQ2 := res["q2"]
	if !foundQ1 || !foundQ2 {
		t.Errorf("Failed to route fanout message")
	}
}

func TestExchangeRoutingTopic(t *testing.T) {
	var exTopic = NewExchange(
		"ext",
		EX_TYPE_TOPIC,
		false,
		false,
		false,
		amqp.NewTable(),
		false,
		make(chan *Exchange),
	)
	exTopic.AddBinding(&amqp.QueueBind{Queue: "q1", Exchange: "ext", RoutingKey: "api.msg.*.json"}, -1, false)
	exTopic.AddBinding(&amqp.QueueBind{Queue: "q1", Exchange: "ext", RoutingKey: "api.*.home.json"}, -1, false)
	exTopic.AddBinding(&amqp.QueueBind{Queue: "q2", Exchange: "ext", RoutingKey: "api.msg.home.json"}, -1, false)
	exTopic.AddBinding(&amqp.QueueBind{Queue: "q3", Exchange: "ext", RoutingKey: "log.#"}, -1, false)

	// Create a random message, won't route by default
	var msg = amqp.RandomMessage(false)
	msg.Method.Exchange = "ext"

	// no match
	res, err := exTopic.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	if len(res) > 0 {
		t.Errorf("Routed message which should not have routed", res)
	}

	// one match on #
	msg.Method.RoutingKey = "log.msg.home.json"
	res, err = exTopic.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	_, foundLog := res["q3"]
	if !foundLog || len(res) != 1 {
		t.Errorf("Bad results routing to # key")
	}

	// one queue on two matches
	msg.Method.RoutingKey = "api.msg.home.json"
	res, err = exTopic.QueuesForPublish(msg)
	if err != nil {
		t.Errorf(err.Msg)
	}
	_, foundQ1 := res["q1"]
	_, foundQ2 := res["q2"]
	if !foundQ1 || !foundQ2 || len(res) != 2 {
		t.Errorf("Bad results routing to multiply-bound * key")
	}

}

func TestPersistence(t *testing.T) {
	// Create DB
	var dbFile = "TestExchangePersistence.db"
	os.Remove(dbFile)
	defer os.Remove(dbFile)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		t.Errorf("Failed to create db")
	}
	err = PersistExchange(db, &amqp.ExchangeDeclare{
		Exchange:   "ex1",
		Type:       "topic",
		Passive:    false,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Arguments:  amqp.NewTable(),
	}, false)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Read
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("exchanges"))
		exBytes := bucket.Get([]byte("ex1"))
		if exBytes == nil {
			return fmt.Errorf("ex1 not found")
		}
		var decl = &amqp.ExchangeDeclare{}
		err = decl.Read(bytes.NewBuffer(exBytes))
		if err != nil {
			return err
		}
		if decl.Type != "topic" {
			return fmt.Errorf("Different exchange types after persisting!")
		}
		return nil
	})
	if err != nil {
		t.Errorf("Error loading persisted exchage %s", err.Error())
	}

	// Depersist
	ex := NewExchange("ex1", EX_TYPE_TOPIC, true, false, false, amqp.NewTable(), false, make(chan *Exchange))
	ex.AddBinding(&amqp.QueueBind{
		Queue:      "q1",
		Exchange:   "ex1",
		RoutingKey: "api.msg.*.json",
		Arguments:  amqp.NewTable(),
	}, -1, false)
	ex.Depersist(db)

	// Verify
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("exchanges"))
		exBytes := bucket.Get([]byte("ex1"))
		if exBytes != nil {
			return fmt.Errorf("ex1 found!")
		}
		return nil
	})
	if err != nil {
		t.Errorf("Error loading persisted exchage %s", err.Error())
	}
}
