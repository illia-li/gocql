package mod

import "reflect"

var IntoRef Mod = func(vals ...interface{}) []interface{} {
	out := make([]interface{}, 0)
	for i := range vals {
		if vals[i] != nil {
			out = append(out, intoRef(vals[i]))
		}
	}
	return out
}

func intoRef(val interface{}) interface{} {
	inV := reflect.ValueOf(val)
	out := reflect.New(reflect.TypeOf(val))
	out.Elem().Set(inV)
	return out.Interface()
}
