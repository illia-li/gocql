package serialization_test

import (
	"github.com/gocql/gocql/serialization/decimal"
	"gopkg.in/inf.v0"
	"math"
	"math/big"
	"testing"

	"github.com/gocql/gocql"
	"github.com/gocql/gocql/internal/tests/serialization"
	"github.com/gocql/gocql/internal/tests/serialization/mod"
)

func TestMarshalDecimal(t *testing.T) {
	tType := gocql.NewNativeType(4, gocql.TypeDecimal, "")

	type testSuite struct {
		name      string
		marshal   func(interface{}) ([]byte, error)
		unmarshal func(bytes []byte, i interface{}) error
	}

	testSuites := [2]testSuite{
		{
			name:      "serialization.decimal",
			marshal:   decimal.Marshal,
			unmarshal: decimal.Unmarshal,
		},
		{
			name: "glob",
			marshal: func(i interface{}) ([]byte, error) {
				return gocql.Marshal(tType, i)
			},
			unmarshal: func(bytes []byte, i interface{}) error {
				return gocql.Unmarshal(tType, bytes, i)
			},
		},
	}

	type scaleCase struct {
		name  string
		scale inf.Scale
		data  []byte
	}

	scales := []scaleCase{
		{name: "scale max", scale: math.MaxInt32, data: []byte("\x7f\xff\xff\xff")},
		{name: "scale 1", scale: 1, data: []byte("\x00\x00\x00\x01")},
		{name: "scale 0", scale: 0, data: []byte("\x00\x00\x00\x00")},
		{name: "scale -1", scale: -1, data: []byte("\xff\xff\xff\xff")},
		{name: "scale min", scale: math.MinInt32, data: []byte("\x80\x00\x00\x00")},
	}

	for _, tSuite := range testSuites {
		marshal := tSuite.marshal
		unmarshal := tSuite.unmarshal

		t.Run(tSuite.name, func(t *testing.T) {

			serialization.PositiveSet{
				Data:   nil,
				Values: mod.Values{(*inf.Dec)(nil)},
			}.Run("[nil]nullable", t, marshal, unmarshal)

			serialization.PositiveSet{
				Data:   nil,
				Values: mod.Values{*inf.NewDec(0, 0)},
			}.Run("[nil]unmarshal", t, nil, unmarshal)

			serialization.PositiveSet{
				Data:   make([]byte, 0),
				Values: mod.Values{*inf.NewDec(0, 0)}.AddVariants(mod.Reference),
			}.Run("[]unmarshal", t, nil, unmarshal)

			serialization.PositiveSet{
				Data:   []byte("\x00\x00\x00\x00\x00"),
				Values: mod.Values{*inf.NewDec(0, 0)}.AddVariants(mod.Reference),
			}.Run("zeros", t, marshal, unmarshal)

			for _, sCase := range scales {
				scale := sCase.scale
				scaleData := sCase.data

				t.Run(sCase.name, func(t *testing.T) {
					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01")...),
						Values: mod.Values{*inf.NewDec(1, scale)}.AddVariants(mod.Reference),
					}.Run("+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff")...),
						Values: mod.Values{*inf.NewDec(-1, scale)}.AddVariants(mod.Reference),
					}.Run("-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f")...),
						Values: mod.Values{*inf.NewDec(127, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt8", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80")...),
						Values: mod.Values{*inf.NewDec(-128, scale)}.AddVariants(mod.Reference),
					}.Run("minInt8", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80")...),
						Values: mod.Values{*inf.NewDec(128, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt8+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f")...),
						Values: mod.Values{*inf.NewDec(-129, scale)}.AddVariants(mod.Reference),
					}.Run("minInt8-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff")...),
						Values: mod.Values{*inf.NewDec(32767, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt16", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00")...),
						Values: mod.Values{*inf.NewDec(-32768, scale)}.AddVariants(mod.Reference),
					}.Run("minInt16", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00")...),
						Values: mod.Values{*inf.NewDec(32768, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt16+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff")...),
						Values: mod.Values{*inf.NewDec(-32769, scale)}.AddVariants(mod.Reference),
					}.Run("minInt16-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff\xff")...),
						Values: mod.Values{*inf.NewDec(8388607, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt24", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00\x00")...),
						Values: mod.Values{*inf.NewDec(-8388608, scale)}.AddVariants(mod.Reference),
					}.Run("minInt24", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00\x00")...),
						Values: mod.Values{*inf.NewDec(8388608, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt24+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff\xff")...),
						Values: mod.Values{*inf.NewDec(-8388609, scale)}.AddVariants(mod.Reference),
					}.Run("minInt24-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(2147483647, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt32", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(-2147483648, scale)}.AddVariants(mod.Reference),
					}.Run("minInt32", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(2147483648, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt32+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(-2147483649, scale)}.AddVariants(mod.Reference),
					}.Run("minInt32-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(549755813887, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt40", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(-549755813888, scale)}.AddVariants(mod.Reference),
					}.Run("minInt40", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(549755813888, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt40+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(-549755813889, scale)}.AddVariants(mod.Reference),
					}.Run("minInt40-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(140737488355327, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt48", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(-140737488355328, scale)}.AddVariants(mod.Reference),
					}.Run("minInt48", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(140737488355328, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt48+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(-140737488355329, scale)}.AddVariants(mod.Reference),
					}.Run("minInt48-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(36028797018963967, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt56", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(-36028797018963968, scale)}.AddVariants(mod.Reference),
					}.Run("minInt56", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(36028797018963968, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt56+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(-36028797018963969, scale)}.AddVariants(mod.Reference),
					}.Run("minInt56-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x7f\xff\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(9223372036854775807, scale)}.AddVariants(mod.Reference),
					}.Run("maxInt64", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x80\x00\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(-9223372036854775808, scale)}.AddVariants(mod.Reference),
					}.Run("minInt64", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\x80\x00\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDecBig(big.NewInt(0).Add(big.NewInt(1), big.NewInt(9223372036854775807)), scale)}.AddVariants(mod.Reference),
					}.Run("maxInt64+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\xff\x7f\xff\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDecBig(big.NewInt(0).Add(big.NewInt(-1), big.NewInt(-9223372036854775808)), scale)}.AddVariants(mod.Reference),
					}.Run("minInt64-1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff")...),
						Values: mod.Values{*inf.NewDec(255, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint8", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00")...),
						Values: mod.Values{*inf.NewDec(256, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint8+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff")...),
						Values: mod.Values{*inf.NewDec(65535, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint16", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00")...),
						Values: mod.Values{*inf.NewDec(65536, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint16+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(16777215, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint24", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(16777216, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint24+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(4294967295, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint32", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(4294967296, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint32+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(1099511627775, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint40", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(1099511627776, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint40+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(281474976710655, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint48", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(281474976710656, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint48+1", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDec(72057594037927935, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint56", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDec(72057594037927936, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint56+1", t, marshal, unmarshal)

					bigMaxUint64 := new(big.Int)
					bigMaxUint64.SetString("18446744073709551615", 10)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x00\xff\xff\xff\xff\xff\xff\xff\xff")...),
						Values: mod.Values{*inf.NewDecBig(bigMaxUint64, scale)}.AddVariants(mod.Reference),
					}.Run("maxUint64", t, marshal, unmarshal)

					serialization.PositiveSet{
						Data:   append(scaleData, []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x00")...),
						Values: mod.Values{*inf.NewDecBig(big.NewInt(0).Add(bigMaxUint64, big.NewInt(1)), scale)}.AddVariants(mod.Reference),
					}.Run("maxUint64+1", t, marshal, unmarshal)
				})
			}
		})
	}
}
