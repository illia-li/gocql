package decimal

import (
	"gopkg.in/inf.v0"
	"math"
)

func EncInfDec(v inf.Dec) ([]byte, error) {
	return encInfDec(v), nil
}

func EncInfDecR(v *inf.Dec) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return encInfDecR(v), nil
}

func encInfDec(v inf.Dec) []byte {
	sign := v.Sign()
	if sign == 0 {
		return []byte{0, 0, 0, 0, 0}
	}

	data := encScale(v.Scale())

	vBig := v.UnscaledBig()
	if sign == 1 {
		dataBig := vBig.Bytes()
		if dataBig[0] > math.MaxInt8 {
			data = append(data, 0)
		}
		return append(data, dataBig...)
	}
	dataBig := vBig.Bytes()
	add := true
	for i := len(dataBig) - 1; i >= 0; i-- {
		if !add {
			dataBig[i] = 255 - dataBig[i]
		} else {
			dataBig[i] = 255 - dataBig[i] + 1
			if dataBig[i] != 0 {
				add = false
			}
		}
	}
	if dataBig[0] < 128 {
		data = append(data, 255)
	}
	return append(data, dataBig...)
}

func encInfDecR(v *inf.Dec) []byte {
	sign := v.Sign()
	if sign == 0 {
		return []byte{0, 0, 0, 0, 0}
	}

	data := encScale(v.Scale())

	vBig := v.UnscaledBig()
	if sign == 1 {
		dataBig := vBig.Bytes()
		if dataBig[0] > math.MaxInt8 {
			data = append(data, 0)
		}
		return append(data, dataBig...)
	}
	dataBig := vBig.Bytes()
	add := true
	for i := len(dataBig) - 1; i >= 0; i-- {
		if !add {
			dataBig[i] = 255 - dataBig[i]
		} else {
			dataBig[i] = 255 - dataBig[i] + 1
			if dataBig[i] != 0 {
				add = false
			}
		}
	}
	if dataBig[0] < 128 {
		data = append(data, 255)
	}
	return append(data, dataBig...)
}

func encScale(v inf.Scale) []byte {
	return []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}
