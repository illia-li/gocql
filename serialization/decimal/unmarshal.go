package decimal

import (
	"fmt"
	"gopkg.in/inf.v0"
)

func Unmarshal(data []byte, value interface{}) error {
	switch v := value.(type) {
	case nil:
		return nil
	case *inf.Dec:
		return DecInfDec(data, v)
	case **inf.Dec:
		return DecInfDecR(data, v)
	default:
		return fmt.Errorf("failed to unmarshal decimal: unsupported value type (%T)(%[1]v)", v)
	}
}
