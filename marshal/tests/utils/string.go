package utils

import (
	"fmt"
	"reflect"
)

const printLimit = 100

func StringPointer(i interface{}) string {
	rv := reflect.ValueOf(i)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		if rv.Elem().Kind() == reflect.Ptr {
			return StringPointer(rv.Elem().Interface())
		}
		return fmt.Sprintf("%d", rv.Pointer())
	}
	return ""
}

func StringData(p []byte) string {
	if len(p) > printLimit {
		p = p[:printLimit]
	}
	return fmt.Sprintf("[%x]", p)
}
