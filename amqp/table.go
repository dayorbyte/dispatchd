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

func (table *Table) SetKey(key string, value interface{}) *FieldValue {
	var fieldValue *FieldValue = nil
	for _, kv := range table.Table {
		var k = *kv.Key
		if k == key {
			value = kv.Value
		}
	}
	if fieldValue == nil {
		fieldValue = &FieldValue{
			Value: nil,
		}
		table.Table = append(table.Table, &FieldValuePair{
			Key:   &key,
			Value: fieldValue,
		})
	}
	fieldValue.Value = calcValue(value)
	return nil
}

func calcValue(value interface{}) isFieldValue_Value {
	switch v := value.(type) {
	case bool:
		return &FieldValue_VBoolean{VBoolean: v}
	case int8:
		return &FieldValue_VInt8{VInt8: v}
	case uint8:
		return &FieldValue_VUint8{VUint8: v}
	case int16:
		return &FieldValue_VInt16{VInt16: v}
	case uint16:
		return &FieldValue_VUint16{VUint16: v}
	case int32:
		return &FieldValue_VInt32{VInt32: v}
	case uint32:
		return &FieldValue_VUint32{VUint32: v}
	case int64:
		return &FieldValue_VInt64{VInt64: v}
	case uint64:
		return &FieldValue_VUint64{VUint64: v}
	case float32:
		return &FieldValue_VFloat{VFloat: v}
	case float64:
		return &FieldValue_VDouble{VDouble: v}
	case *Decimal:
		return &FieldValue_VDecimal{VDecimal: v}
	case string:
		return &FieldValue_VShortstr{VShortstr: v}
	case []byte:
		return &FieldValue_VLongstr{VLongstr: v}
	case *FieldArray:
		return &FieldValue_VArray{VArray: v}
	// TODO: not currently reachable since uint64 will take it
	// case timestamp:
	//   return &FieldValue_VTimestamp{VTimestamp: v}
	case *Table:
		return &FieldValue_VTable{VTable: v}
	}
	return nil
}
