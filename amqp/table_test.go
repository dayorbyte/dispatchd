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

func TestTableTypes(t *testing.T) {
	var inTable = NewTable()

	// Basic types
	inTable.SetKey("bool", true)
	inTable.SetKey("int8", int8(-2))
	inTable.SetKey("uint8", uint8(3))
	inTable.SetKey("int16", int16(-4))
	inTable.SetKey("uint16", uint16(5))
	inTable.SetKey("int32", int32(-6))
	inTable.SetKey("uint32", uint32(7))
	inTable.SetKey("int64", int64(-8))
	inTable.SetKey("uint64", uint64(9))
	inTable.SetKey("float32", float32(10.1))
	inTable.SetKey("float64", float64(-11.2))
	inTable.SetKey("string", "string value")
	inTable.SetKey("[]byte", []byte{14, 15, 16, 17})
	// Decimal
	var scale = uint8(12)
	var value = int32(-13)
	inTable.SetKey("*Decimal", &Decimal{&scale, &value})
	// Field Array
	var fa = NewFieldArray()
	fa.AppendFA(int8(101))
	inTable.SetKey("*FieldArray", fa)
	// Table
	var innerTable = NewTable()
	innerTable.SetKey("some key", "some value")
	inTable.SetKey("*Table", innerTable)

	// TODO: timestamp

	// Encode
	writer := bytes.NewBuffer(make([]byte, 0))
	err := WriteTable(writer, inTable)
	if err != nil {
		t.Errorf(err.Error())
	}

	// decode
	var reader = bytes.NewReader(writer.Bytes())
	outTable, err := ReadTable(reader)
	if err != nil {
		t.Errorf(err.Error())
	}

	// compare
	if !EquivalentTables(inTable, outTable) {
		t.Errorf("Tables no equal")
	}

}
