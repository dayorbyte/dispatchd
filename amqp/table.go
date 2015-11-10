package amqp

import (
	"reflect"
)

func NewTable() *Table {
	return &Table{Table: make([]*FieldValuePair, 0)}
}

func EquivalentTables(t1 *Table, t2 *Table) bool {
	return reflect.DeepEqual(t1, t2)
}

func (table *Table) GetKey(key string) *FieldValue {
	for _, kv := range table.Table {
		var k = *kv.Key
		if k == key {
			return kv.Value
		}
	}
	return nil
}
