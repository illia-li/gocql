package utils

import (
	"reflect"
)

func DeReference(in interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(in)).Interface()
}
