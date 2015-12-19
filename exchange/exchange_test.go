package exchange

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/dispatchd/amqp"
	"github.com/jeffjenkins/dispatchd/binding"
	"os"
	"reflect"
	"testing"
	"time"
)

func exchangeForTest(name string, typ uint8) *Exchange {
	return NewExchange(
		name,
		typ,
		false,
		false,
		false,
		amqp.NewTable(),
		false,
		make(chan *Exchange),
	)
}

func TestClose(t *testing.T) {
	var ex = exchangeForTest("ex", EX_TYPE_TOPIC)
	if ex.Closed {
		t.Errorf("Exchange closed when it shouldn't be!")
	}
	ex.Close()
	if !ex.Closed {
		t.Errorf("Exchange closed when it shouldn't be!")
	}
}

func TestJSON(t *testing.T) {
	var ex = exchangeForTest("ex", EX_TYPE_TOPIC)
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
	ex.ExType = 123
	_, err = json.Marshal(ex)
	if err == nil {
		t.Errorf("Didn't get error in json.Marshal with an invalid extype")
	}
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
	if _, err := exchangeTypeToName(123); err == nil {
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
	ex2.ExType = EX_TYPE_TOPIC
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.ExType = EX_TYPE_DIRECT
	// internal
	ex2.Internal = false
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.Internal = true
	// durable
	ex2.Durable = false
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.Durable = true
	// args
	var newTable = amqp.NewTable()
	newTable.SetKey("stuff", true)
	ex2.Arguments = newTable
	if ex.EquivalentExchanges(ex2) {
		t.Errorf("Different exchanges are equal!")
	}
	ex2.Arguments = amqp.NewTable()
	// test other diffs ok
	ex2.System = false
	ex2.AutoDelete = false
	if !ex.EquivalentExchanges(ex2) {
		t.Errorf("Same exchanges aren't equal!")
	}
}

func bindingHelper(queue string, exchange string, key string, topic bool) *binding.Binding {
	b, err := binding.NewBinding(queue, exchange, key, amqp.NewTable(), topic)
	if err != nil {
		panic(err.Error())
	}
	return b
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
	exDirect.AddBinding(bindingHelper("q1", "exd", "rk-1", false), -1)

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
	exFanout.AddBinding(bindingHelper("q1", "exf", "rk-1", false), -1)
	exFanout.AddBinding(bindingHelper("q2", "exf", "rk-2", false), -1)

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
		t.Errorf("Failed to route fanout message %v %v", foundQ1, foundQ2)
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
	exTopic.AddBinding(bindingHelper("q1", "ext", "api.msg.*.json", true), -1)
	exTopic.AddBinding(bindingHelper("q1", "ext", "api.*.home.json", true), -1)
	exTopic.AddBinding(bindingHelper("q2", "ext", "api.msg.home.json", true), -1)
	exTopic.AddBinding(bindingHelper("q3", "ext", "log.#", true), -1)

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
	var ex = exchangeForTest("ex1", EX_TYPE_TOPIC)
	ex.Durable = true
	err = ex.Persist(db)
	if err != nil {
		t.Errorf("Could not persist exchange %s", ex.Name)
	}
	ex.Name = ""
	err = ex.Persist(db)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Read
	deleteChan := make(chan *Exchange)
	_, err = NewFromDisk(db, "ex1", deleteChan)
	if err != nil {
		t.Errorf("Error loading persisted exchage %s", err.Error())
	}
	_, err = NewFromDisk(db, "", deleteChan)
	if err != nil {
		t.Errorf("Error loading persisted exchage %s", err.Error())
	}

	// Depersist
	realEx := NewExchange("ex1", EX_TYPE_TOPIC, true, false, false, amqp.NewTable(), false, make(chan *Exchange))
	realEx.AddBinding(bindingHelper("q1", "ex1", "api.msg.*.json", true), -1)
	realEx.Depersist(db)

	// Verify
	_, err = NewFromDisk(db, "ex1", deleteChan)
	if err == nil {
		t.Errorf("Failed to delete exchange 'ex1'")
	}
}

func TestAddBinding(t *testing.T) {
	var ex = NewExchange("ex1", EX_TYPE_TOPIC, true, true, false, amqp.NewTable(), false, make(chan *Exchange))
	ex.deleteActive = time.Now()
	// bad binding
	_, err := binding.NewBinding("q1", "ex1", "~!@", amqp.NewTable(), true)
	if err == nil {
		t.Errorf("No error with bad binding!")
	}
	if len(ex.bindings) != 0 {
		t.Errorf("Bad binding was added despite error")
	}
	// duplicate binding
	var b = bindingHelper("q1", "ex1", "a.b.c", true)
	err = ex.AddBinding(b, -1)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(ex.bindings) != 1 {
		t.Errorf("Wrong number of bindings")
	}
	err = ex.AddBinding(b, -1)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(ex.bindings) != 1 {
		t.Errorf("Wrong number of bindings")
	}
	if ex.deleteActive != time.Unix(0, 0) {
		t.Errorf("Error did not reset time")
	}

}

func TestBindingsForQueue(t *testing.T) {
	var ex = NewExchange("ex1", EX_TYPE_TOPIC, true, true, false, amqp.NewTable(), false, make(chan *Exchange))
	var b1 = bindingHelper("q1", "ex1", "a.b.c", true)
	var b2 = bindingHelper("q1", "ex1", "d.e.f", true)
	var b3 = bindingHelper("q1", "ex1", "g.h.i", true)
	var b4 = bindingHelper("q2", "ex1", "g.h.i", true)
	var b5 = bindingHelper("q2", "ex1", "j.k.l", true)
	//
	ex.AddBinding(b1, -1)
	ex.AddBinding(b2, -1)
	ex.AddBinding(b3, -1)
	ex.AddBinding(b4, -1)
	ex.AddBinding(b5, -1)

	if len(ex.BindingsForQueue("q1")) != 3 {
		t.Errorf("Wrong number of bindings for q1")
	}
	if len(ex.BindingsForQueue("q2")) != 2 {
		t.Errorf("Wrong number of bindings for q2")
	}
	if len(ex.BindingsForQueue("q0")) != 0 {
		t.Errorf("Wrong number of bindings for q0")
	}
}

func TestRemoveBindingsForQueue(t *testing.T) {
	var ex = NewExchange("ex1", EX_TYPE_TOPIC, true, true, false, amqp.NewTable(), false, make(chan *Exchange))
	var b1 = bindingHelper("q1", "ex1", "a.b.c", true)
	var b2 = bindingHelper("q1", "ex1", "d.e.f", true)
	var b3 = bindingHelper("q1", "ex1", "g.h.i", true)
	var b4 = bindingHelper("q2", "ex1", "g.h.i", true)
	var b5 = bindingHelper("q2", "ex1", "j.k.l", true)
	//
	ex.AddBinding(b1, -1)
	ex.AddBinding(b2, -1)
	ex.AddBinding(b3, -1)
	ex.AddBinding(b4, -1)
	ex.AddBinding(b5, -1)

	ex.RemoveBindingsForQueue("q0")
	if len(ex.bindings) != 5 {
		t.Errorf("Wrong number of bindings after removing q0 bindings")
	}
	ex.RemoveBindingsForQueue("q1")
	if len(ex.bindings) != 2 {
		t.Errorf("Wrong number of bindings after removing q1 bindings")
	}
	ex.RemoveBindingsForQueue("q2")
	if len(ex.bindings) != 0 {
		t.Errorf("Wrong number of bindings after removing q2 bindings: %v", ex.bindings)
	}

}

// func NewBinding(queueName string, exchangeName string, key string, arguments *amqp.Table, topic bool) (*Binding, error) {

func TestRemoveBinding(t *testing.T) {
	var ex = NewExchange("ex1", EX_TYPE_TOPIC, true, true, false, amqp.NewTable(), false, make(chan *Exchange))
	var b1 = bindingHelper("q1", "ex1", "a.b.c", true)
	var b2 = bindingHelper("q1", "ex1", "d.e.f", true)
	var b3 = bindingHelper("q1", "ex1", "g.h.i", true)
	var b4 = bindingHelper("q2", "ex1", "g.h.i", true)
	var b5 = bindingHelper("q2", "ex1", "j.k.l", true)
	//
	ex.AddBinding(b1, -1)
	ex.AddBinding(b2, -1)
	ex.AddBinding(b3, -1)
	ex.AddBinding(b4, -1)
	ex.AddBinding(b5, -1)

	// Remove a binding that doesn't exist
	bNone := bindingHelper("q2", "ex1", "does.not.exist", true)
	ex.RemoveBinding(bNone)
	if len(ex.bindings) != 5 {
		t.Errorf("Wrong number of bindings: %d", len(ex.bindings))
	}

	// Remove the Q2 bindings
	ex.RemoveBinding(b4)
	if len(ex.bindings) != 4 {
		t.Errorf("Wrong number of bindings: %d", len(ex.bindings))
	}
	ex.RemoveBinding(b5)

	// Check that all q2 bindings are gone
	if len(ex.BindingsForQueue("q2")) != 0 {
		t.Errorf("Wrong number of bindings")
	}

}

func TestAutoDeleteTimeout(t *testing.T) {
	var deleter = make(chan *Exchange)
	var ex = NewExchange("ex1", EX_TYPE_TOPIC, true, true, false, amqp.NewTable(), false, deleter)
	ex.autodeletePeriod = 10 * time.Millisecond
	ex.AddBinding(bindingHelper("q1", "ex1", "a.b.c", true), -1)
	var b, _ = binding.NewBinding("q1", "ex1", "a.b.c", amqp.NewTable(), true)
	ex.RemoveBinding(b)
	var toDelete = <-deleter
	if ex.Name != toDelete.Name {
		t.Errorf("Integrity error in delete")
	}
}

func TestNoExchangesBucket(t *testing.T) {
	var dbFile = "TestNoExchangeBucket.db"
	os.Remove(dbFile)
	defer os.Remove(dbFile)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		t.Errorf("Failed to create db")
	}
	_, err = NewFromDisk(db, "does-not-exist", make(chan *Exchange))
	if err == nil {
		t.Errorf("No error trying to load an non-existent bucket")
	}
}

func TestBadExchangeBytesBucket(t *testing.T) {
	var dbFile = "TestBadExchangeBytes.db"
	os.Remove(dbFile)
	defer os.Remove(dbFile)
	db, err := bolt.Open(dbFile, 0600, nil)
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("exchanges"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return bucket.Put([]byte("bad-bytes"), []byte{0, 1, 2, 3})
	})
	if err != nil {
		t.Errorf(err.Error())
	}
	NewFromDisk(db, "bad-bytes", make(chan *Exchange))
}

func TestNewFromMethod(t *testing.T) {
	var method = &amqp.ExchangeDeclare{
		Reserved1:  0,
		Exchange:   "ex-method",
		Type:       "topic",
		Passive:    true,
		Durable:    true,
		AutoDelete: true,
		Internal:   true,
		NoWait:     false,
		Arguments:  amqp.NewTable(),
	}
	exMethod, err := NewFromMethod(method, true, make(chan *Exchange))
	if err != nil {
		t.Errorf(err.Msg)
	}
	exNormal := NewExchange("ex-method", EX_TYPE_TOPIC, true, true, true, amqp.NewTable(), true, make(chan *Exchange))
	if !exNormal.EquivalentExchanges(exMethod) {
		t.Errorf("Inconsistency between NewExchange and NewFromMethod")
	}
	// Bad exchange type
	method.Type = "headers"
	exMethod, err = NewFromMethod(method, true, make(chan *Exchange))
	if err == nil {
		t.Errorf("Parsed bad exchange method")
	}
	if err.Code != 503 {
		t.Errorf("Wrong error code on bad exchange parse")
	}
}
