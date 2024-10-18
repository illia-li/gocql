package decimal

import (
	"fmt"
	"gopkg.in/inf.v0"
	"math/big"
)

var errWrongDataLen = fmt.Errorf("failed to unmarshal decimal: the length of the data should be 0 or more than 5")

func errBrokenData(p []byte) error {
	if p[4] == 0 && p[5] <= 127 || p[4] == 255 && p[5] > 127 {
		return fmt.Errorf("failed to unmarshal decimal: the data is broken")
	}
	return nil
}

func errNilReference(v interface{}) error {
	return fmt.Errorf("failed to unmarshal decimal: can not unmarshal into nil reference(%T)(%[1]v)", v)
}

func DecInfDec(p []byte, v *inf.Dec) error {
	if v == nil {
		return errNilReference(v)
	}
	switch len(p) {
	case 0:
		v.SetScale(0)
		v.SetUnscaled(0)
		return nil
	case 1, 2, 3, 4:
		return errWrongDataLen
	case 5:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec1toInt64(p))
		return nil
	case 6:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec2toInt64(p))
	case 7:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec3toInt64(p))
	case 8:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec4toInt64(p))
	case 9:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec5toInt64(p))
	case 10:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec6toInt64(p))
	case 11:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec7toInt64(p))
	case 12:
		v.SetScale(decScale(p))
		v.SetUnscaled(dec8toInt64(p))
	default:
		v.SetScale(decScale(p))
		v.SetUnscaledBig(dec2BigInt(p[4:]))
	}
	return errBrokenData(p)
}

func DecInfDecR(p []byte, v **inf.Dec) error {
	if v == nil {
		return errNilReference(v)
	}
	switch len(p) {
	case 0:
		if p == nil {
			*v = nil
		} else {
			*v = inf.NewDec(0, 0)
		}
		return nil
	case 1, 2, 3, 4:
		return errWrongDataLen
	case 5:
		*v = inf.NewDec(dec1toInt64(p), decScale(p))
		return nil
	case 6:
		*v = inf.NewDec(dec2toInt64(p), decScale(p))
	case 7:
		*v = inf.NewDec(dec3toInt64(p), decScale(p))
	case 8:
		*v = inf.NewDec(dec4toInt64(p), decScale(p))
	case 9:
		*v = inf.NewDec(dec5toInt64(p), decScale(p))
	case 10:
		*v = inf.NewDec(dec6toInt64(p), decScale(p))
	case 11:
		*v = inf.NewDec(dec7toInt64(p), decScale(p))
	case 12:
		*v = inf.NewDec(dec8toInt64(p), decScale(p))
	default:
		*v = inf.NewDecBig(dec2BigInt(p[4:]), decScale(p))
	}
	return errBrokenData(p)
}

func decScale(p []byte) inf.Scale {
	return inf.Scale(p[0])<<24 | inf.Scale(p[1])<<16 | inf.Scale(p[2])<<8 | inf.Scale(p[3])
}

func dec1toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<8 | int64(p[4])
	}
	return int64(p[4])
}

func dec2toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<16 | int64(p[4])<<8 | int64(p[5])
	}
	return int64(p[4])<<8 | int64(p[5])
}

func dec3toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<24 | int64(p[4])<<16 | int64(p[5])<<8 | int64(p[6])
	}
	return int64(p[4])<<16 | int64(p[5])<<8 | int64(p[6])
}

func dec4toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<32 | int64(p[4])<<24 | int64(p[5])<<16 | int64(p[6])<<8 | int64(p[7])
	}
	return int64(p[4])<<24 | int64(p[5])<<16 | int64(p[6])<<8 | int64(p[7])
}

func dec5toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<40 | int64(p[4])<<32 | int64(p[5])<<24 | int64(p[6])<<16 | int64(p[7])<<8 | int64(p[8])
	}
	return int64(p[4])<<32 | int64(p[5])<<24 | int64(p[6])<<16 | int64(p[7])<<8 | int64(p[8])
}

func dec6toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<48 | int64(p[4])<<40 | int64(p[5])<<32 | int64(p[6])<<24 | int64(p[7])<<16 | int64(p[8])<<8 | int64(p[9])
	}
	return int64(p[4])<<40 | int64(p[5])<<32 | int64(p[6])<<24 | int64(p[7])<<16 | int64(p[8])<<8 | int64(p[9])
}

func dec7toInt64(p []byte) int64 {
	if p[4] > 127 {
		return int64(-1)<<56 | int64(p[4])<<48 | int64(p[5])<<40 | int64(p[6])<<32 | int64(p[7])<<24 | int64(p[8])<<16 | int64(p[9])<<8 | int64(p[10])
	}
	return int64(p[4])<<48 | int64(p[5])<<40 | int64(p[6])<<32 | int64(p[7])<<24 | int64(p[8])<<16 | int64(p[9])<<8 | int64(p[10])
}

func dec8toInt64(p []byte) int64 {
	return int64(p[4])<<56 | int64(p[5])<<48 | int64(p[6])<<40 | int64(p[7])<<32 | int64(p[8])<<24 | int64(p[9])<<16 | int64(p[10])<<8 | int64(p[11])
}

func dec2BigInt(p []byte) *big.Int {
	// Positive range processing
	if p[0] <= 127 {
		return new(big.Int).SetBytes(p)
	}
	// negative range processing
	data := make([]byte, len(p))
	copy(data, p)

	add := true
	for i := len(data) - 1; i >= 0; i-- {
		if !add {
			data[i] = 255 - data[i]
		} else {
			data[i] = 255 - data[i] + 1
			if data[i] != 0 {
				add = false
			}
		}
	}

	return new(big.Int).Neg(new(big.Int).SetBytes(data))
}
