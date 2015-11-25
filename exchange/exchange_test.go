package exchange

import (
	"github.com/jeffjenkins/mq/amqp"
	// "github.com/jeffjenkins/mq/binding"
	// "github.com/jeffjenkins/mq/queue"
	"testing"
)

// EX_TYPE_DIRECT
// EX_TYPE_FANOUT
// EX_TYPE_TOPIC
// EX_TYPE_HEADERS

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
	if _, err := exchangeTypeToName(Extype(12345)); err != nil {
		t.Errorf(err.Error())
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
