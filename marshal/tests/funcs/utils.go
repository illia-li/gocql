package funcs

import (
	"fmt"
	"reflect"
)

var (
	ExcludedMarshal   = func(interface{}) ([]byte, error) { return nil, fmt.Errorf("run on ExcludedMarshal func") }
	ExcludedUnmarshal = func([]byte, interface{}) error { return fmt.Errorf("run on ExcludedUnmarshal func") }
)

func IsExcludedMarshal(f func(interface{}) ([]byte, error)) bool {
	return reflect.ValueOf(f).Pointer() == reflect.ValueOf(ExcludedMarshal).Pointer()
}

func IsExcludedUnmarshal(f func([]byte, interface{}) error) bool {
	return reflect.ValueOf(f).Pointer() == reflect.ValueOf(ExcludedUnmarshal).Pointer()
}
