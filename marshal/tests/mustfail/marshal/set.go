package marshal

import (
	"errors"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/gocql/gocql/marshal/tests/mod"
	"github.com/gocql/gocql/marshal/tests/utils"
)

type Set struct {
	MarshalIns []interface{}
	Issue      string
}

func (s Set) AddModified(mods ...mod.Mod) Set {
	out := s
	for _, m := range mods {
		out.MarshalIns = append(out.MarshalIns, m.Apply(s.MarshalIns)...)
	}
	return out
}

func (s Set) Run(name string, t *testing.T, f func(interface{}) ([]byte, error)) {
	t.Logf("test set %s started", name)
	for m := range s.MarshalIns {
		val := s.MarshalIns[m]

		if s.Issue != "" {
			t.Logf("\nskipped bacause there is unsolved issue:\n%s", s.Issue)
			return
		}

		data, err := func() (d []byte, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = utils.PanicErr{Err: r.(error), Stack: debug.Stack()}
				}
			}()
			return f(val)
		}()

		// Prepare test info message
		infoVal := utils.StringValue(val)
		info := ""
		if len(infoVal) < utils.printLimit && len(data) < utils.printLimit {
			info = fmt.Sprintf("\nvalue:%s\nreceived data:%x", infoVal, utils.StringData(data))
		}

		if err == nil {
			t.Errorf("for (%T) marshal does not return error%s", val, info)
		} else if errors.As(err, &utils.PanicErr{}) {
			t.Errorf("for (%T) %s\nwas panic: %s", val, info, err)
		}
	}
	t.Logf("test set %s finished", name)
}
