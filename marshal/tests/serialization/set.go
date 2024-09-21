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
	t.Run(name, func(t *testing.T) {
		for i := range s.Values {
			val := s.Values[i]

			t.Run(fmt.Sprintf("%T", val), func(t *testing.T) {
				if marshal != nil {
					s.runMarshal(t, marshal, val)
				}

				if unmarshal != nil {
					s.runUnmarshal(t, unmarshal, val)
				}
			})
		}
	})
}

func (s Set) runMarshal(t *testing.T, f func(interface{}) ([]byte, error), val interface{}) {
	t.Run("marshal", func(tt *testing.T) {
		if s.IssueMarshal != "" {
			tt.Skipf("skipped bacause there is unsolved issue: %s", s.IssueMarshal)
		}

		result, err := func() (d []byte, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = utils.PanicErr{Err: r.(error), Stack: debug.Stack()}
				}
			}()
			return f(val)
		}()

		if err != nil {
			tt.Fatalf("marshal unexpectedly failed with error: %w", err)
		}

		if !funcs.EqualData(s.Data, result) {
			tt.Errorf("expect %s but got %s", utils.StringData(s.Data), utils.StringData(result))
		}
	})
}

func (s Set) runUnmarshal(t *testing.T, f func([]byte, interface{}) error, expected interface{}) {
	t.Run("unmarshal", func(tt *testing.T) {
		if s.IssueUnmarshal != "" {
			t.Skipf("skipped bacause there is unsolved issue: %s", s.IssueUnmarshal)
		}

		result := funcs.New(expected)
		inValPtr := utils.StringPointer(result)

		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = utils.PanicErr{Err: r.(error), Stack: debug.Stack()}
				}
			}()
			return f(bytes.Clone(s.Data), result)
		}()

		if err != nil {
			tt.Fatalf("unmarshal unexpectedly failed with error: %w", err)
		}

		if outValPtr := utils.StringPointer(result); inValPtr != "" && outValPtr != "" && inValPtr != outValPtr {
			tt.Fatalf("for (%T) unmarshal function rewrites existing pointer", expected)
		}

		if !funcs.EqualVals(expected, utils.DeReference(result)) {
			tt.Errorf("expect %s but got %s", utils.StringValue(expected), utils.StringValue(result))
		}
	})
}
