package serialization

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/gocql/gocql/marshal/tests/funcs"
	"github.com/gocql/gocql/marshal/tests/mod"
	"github.com/gocql/gocql/marshal/tests/utils"
)

type Sets []*Set

// Set is a tool for generating test cases of marshal and unmarshall funcs.
// For cases when the function should no error,
// marshaled data from Set.Values should be equal with Set.Data,
// unmarshalled value from Set.Data should be equal with Set.Values.
type Set struct {
	Data   []byte
	Values []interface{}

	IssueMarshal   string
	IssueUnmarshal string
}

func (s Set) AddModified(mods ...mod.Mod) Set {
	out := s
	for _, m := range mods {
		out.Values = append(out.Values, m.Apply(s.Values)...)
	}
	return out
}

func (s Set) Run(name string, t *testing.T, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) {
	t.Logf("test set %s started", name)
	for i := range s.Values {
		val := s.Values[i]

		if !funcs.IsExcludedMarshal(marshal) {
			s.runMarshal(t, marshal, val)
		}

		if !funcs.IsExcludedUnmarshal(unmarshal) {
			s.runUnmarshal(t, unmarshal, val)
		}
	}
	t.Logf("test set %s finished", name)
}

func (s Set) runMarshal(t *testing.T, f func(interface{}) ([]byte, error), val interface{}) {
	if s.IssueMarshal != "" {
		t.Logf("\nmarshal test skipped bacause there is unsolved issue:\n%s", s.IssueMarshal)
		return
	}

	received, err := func() (d []byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = utils.PanicErr{Err: r.(error), Stack: debug.Stack()}
			}
		}()
		return f(val)
	}()

	info := ""
	if inStr := valStr(val); len(inStr)+len(s.Data)+len(received) < utils.PrintLimit*3 {
		info = fmt.Sprintf("\n marshal   in:%s\nexpected data:%s\nreceived data:%s", valStr(val), dataStr(s.Data), dataStr(received))
	}

	switch {
	case err != nil:
		t.Errorf("for (%T) was error:%s", val, err)
	case !funcs.EqualData(s.Data, received):
		t.Errorf("for (%T) expected and received data are not equal%s", val, info)
	default:
		t.Logf("for (%T) test done%s", val, info)
	}
}

func (s Set) runUnmarshal(t *testing.T, f func([]byte, interface{}) error, val interface{}) {
	if s.IssueUnmarshal != "" {
		t.Logf("\nunmarshal test skipped bacause there is unsolved issue:\n%s", s.IssueUnmarshal)
		return
	}

	inputVal := funcs.New(val)
	inValStr := valStr(inputVal)
	inValPtr := ptrStr(inputVal)

	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = utils.PanicErr{Err: r.(error), Stack: debug.Stack()}
			}
		}()
		return f(bytes.Clone(s.Data), inputVal)
	}()
	if err != nil {
		t.Errorf("for (%T) was error:%s", val, err)
		return
	}

	outValStr := valStr(deRef(inputVal))
	if outValPtr := ptrStr(inputVal); inValPtr != "" && outValPtr != "" && inValPtr != outValPtr {
		t.Errorf("for (%T) unmarshal function rewrites existing pointer", val)
		return
	}

	result := ""
	if expectedStr := valStr(val); len(expectedStr)+len(inValStr)+len(outValStr) < utils.PrintLimit*3 {
		result = fmt.Sprintf("\n     expected:%s\nunmarshal  in:%s\nunmarshal out:%s", expectedStr, inValStr, outValStr)
	}

	if !funcs.EqualVals(val, deRef(inputVal)) {
		t.Errorf("for (%T) expected and received values are not equal%s", val, result)
	} else {
		t.Logf("for (%T) test done%s", val, result)
	}
}
