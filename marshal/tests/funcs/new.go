package funcs

import (
	"reflect"
)

func New(in interface{}) interface{} {
	return getNew(reflect.ValueOf(in))
}

func getNew(orig reflect.Value) reflect.Value {
	if orig.Kind() != reflect.Ptr {
		return reflect.Zero(orig.Type())
	}
	if orig.IsNil() {
		return reflect.Zero(orig.Type())
	} else if orig.IsZero() {
		return reflect.Zero(orig.Type())
	}
	newVal := reflect.New(orig.Type().Elem())
	newVal.Elem().Set(getNew(orig.Elem()))
	return newVal
}
