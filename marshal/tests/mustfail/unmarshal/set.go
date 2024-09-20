package unmarshal

import (
	"bytes"
	"errors"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/gocql/gocql/marshal/tests/mod"
	"github.com/gocql/gocql/marshal/tests/utils"
)

type Set struct {
	Data         []byte
	UnmarshalIns []interface{}

	Issue string
}

func (s Set) AddModified(mods ...mod.Mod) Set {
	out := s
	for _, m := range mods {
		out.UnmarshalIns = append(out.UnmarshalIns, m.Apply(s.UnmarshalIns)...)
	}
	return out
}

func (s Set) Run(name string, t *testing.T, f func([]byte, interface{}) error) {
	t.Logf("test set %s started", name)
	for i := range s.UnmarshalIns {
		val := s.UnmarshalIns[i]
		data := bytes.Clone(s.Data)

		if s.Issue != "" {
			t.Logf("\nskipped bacause there is unsolved issue:\n%s", s.Issue)
			return
		}
		infoIn := utils.StringValue(deRef(val))

		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = utils.PanicErr{Err: r.(error), Stack: debug.Stack()}
				}
			}()
			return f(data, val)
		}()

		// Prepare test info message
		infoOut := utils.StringValue(deRef(val))
		info := ""
		if len(infoOut) < utils.PrintLimit && len(infoIn) < utils.PrintLimit || len(data) < utils.PrintLimit {
			printData := utils.StringData(data)
			info = fmt.Sprintf("\n  tested data:%s\nunmarshal  in:%s\nunmarshal out:%s", printData, infoIn, infoOut)
		}

		if err == nil {
			t.Errorf("for (%T) unmarshal does not return an error%s", val, infoOut)
		} else if errors.As(err, &utils.PanicErr{}) {
			t.Errorf("for (%T) %s\nwas panic: %s", val, info, err)
		} else {
			t.Logf("for (%T) test done%s", val, info)
		}
	}
	t.Logf("test set %s finished", name)
}

func deRef(in interface{}) interface{} {
	return utils.DeReference(in)
}
