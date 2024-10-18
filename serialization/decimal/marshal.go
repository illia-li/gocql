package decimal

import (
	"fmt"
	"gopkg.in/inf.v0"
)

func Marshal(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case inf.Dec:
		return EncInfDec(v)
	case *inf.Dec:
		return EncInfDecR(v)
	default:
		return nil, fmt.Errorf("failed to marshal decimal: unsupported value type (%T)(%[1]v)", v)
	}
}
