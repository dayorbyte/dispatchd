package amqp

import (
	"bytes"
	"testing"
)

func TestTableRoundtrip(t *testing.T) {
	// Create
	var inTable = NewTable()
	inTable.SetKey("product", "mq")
	inTable.SetKey("version", "0.1")
	inTable.SetKey("copyright", "Jeffrey Jenkins, 2015")
	// encode
	writer := bytes.NewBuffer(make([]byte, 0))
	err := WriteTable(writer, inTable)
	if err != nil {
		t.Errorf(err.Error())
	}

	var reader = bytes.NewReader(writer.Bytes())
	outTable, err := ReadTable(reader)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !EquivalentTables(inTable, outTable) {
		t.Errorf("Tables no equal")
	}
}
