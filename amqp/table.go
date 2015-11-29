package amqp

import (
	"fmt"
	"reflect"
)

func NewTable() *Table {
	return &Table{Table: make([]*FieldValuePair, 0)}
}

func EquivalentTables(t1 *Table, t2 *Table) bool {
	if len(t1.Table) == 0 && len(t2.Table) == 0 {
		return true
	}
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

func (table *Table) SetKey(key string, value interface{}) error {
	var fieldValue *FieldValue = nil
	for _, kv := range table.Table {
		var k = *kv.Key
		if k == key {
			fieldValue = kv.Value
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
	fv, err := calcValue(value)

	if err != nil {
		return err
	}
	fieldValue.Value = fv
	return nil
}

func calcValue(value interface{}) (isFieldValue_Value, error) {
	switch v := value.(type) {
	case bool:
		return &FieldValue_VBoolean{VBoolean: v}, nil
	case int8:
		return &FieldValue_VInt8{VInt8: v}, nil
	case uint8:
		return &FieldValue_VUint8{VUint8: v}, nil
	case int16:
		return &FieldValue_VInt16{VInt16: v}, nil
	case uint16:
		return &FieldValue_VUint16{VUint16: v}, nil
	case int32:
		return &FieldValue_VInt32{VInt32: v}, nil
	case uint32:
		return &FieldValue_VUint32{VUint32: v}, nil
	case int64:
		return &FieldValue_VInt64{VInt64: v}, nil
	case uint64:
		return &FieldValue_VUint64{VUint64: v}, nil
	case float32:
		return &FieldValue_VFloat{VFloat: v}, nil
	case float64:
		return &FieldValue_VDouble{VDouble: v}, nil
	case *Decimal:
		return &FieldValue_VDecimal{VDecimal: v}, nil
	case string:
		return &FieldValue_VShortstr{VShortstr: v}, nil
	case []byte:
		return &FieldValue_VLongstr{VLongstr: v}, nil
	case *FieldArray:
		return &FieldValue_VArray{VArray: v}, nil
	// TODO: not currently reachable since uint64 will take it
	// case timestamp:
	//   return &FieldValue_VTimestamp{VTimestamp: v}
	case *Table:
		return &FieldValue_VTable{VTable: v}, nil
	}
	return nil, fmt.Errorf("Field value has invalid type for table/array: %s", value)
}

func NewFieldArray() *FieldArray {
	return &FieldArray{Value: make([]*FieldValue, 0)}
}

func (fa *FieldArray) AppendFA(value interface{}) error {
	var fv, err = calcValue(value)
	if err != nil {
		return err
	}
	fa.Value = append(fa.Value, &FieldValue{Value: fv})
	return nil
}
