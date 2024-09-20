package gocql

import (
	"testing"

	"github.com/gocql/gocql/marshal/tests/funcs"
	"github.com/gocql/gocql/marshal/tests/mod"
	"github.com/gocql/gocql/marshal/tests/serialization"
)

func TestMarshalSmallint(t *testing.T) {
	marshal := func(i interface{}) ([]byte, error) { return Marshal(NativeType{proto: 4, typ: TypeSmallInt}, i) }
	unmarshal := func(bytes []byte, i interface{}) error {
		return Unmarshal(NativeType{proto: 4, typ: TypeSmallInt}, bytes, i)
	}

	serialization.Set{
		Data:   nil,
		Values: []interface{}{(*string)(nil)},
	}.Run("[nil]str", t, marshal, unmarshal)

	serialization.Set{
		Data:   []byte("\x00\x00"),
		Values: []interface{}{"0"},
	}.AddModified(mod.IntoRef).Run("[0000]str", t, marshal, unmarshal)

	serialization.Set{
		Data:   []byte("\x7f\xff"),
		Values: []interface{}{"32767"},
	}.AddModified(mod.IntoRef).Run("[7fff]str", t, marshal, unmarshal)

	serialization.Set{
		Data:   []byte("\x80\x00"),
		Values: []interface{}{"-32768"},
	}.AddModified(mod.IntoRef).Run("[8000]str", t, marshal, unmarshal)

	serialization.Set{
		Data: nil,
		Values: []interface{}{
			(*int8)(nil), (*int16)(nil), (*int32)(nil), (*int64)(nil), (*int)(nil),
			(*uint8)(nil), (*uint16)(nil), (*uint32)(nil), (*uint64)(nil), (*uint)(nil)},
	}.AddModified(mod.IntoCustom).Run("[nil]refs", t, marshal, unmarshal)

	serialization.Set{
		Data: nil,
		Values: []interface{}{
			int8(0), int16(0), int32(0), int64(0), int(0),
			uint8(0), uint16(0), uint32(0), uint64(0), uint(0),
		},
	}.AddModified(mod.IntoCustom).Run("unmarshal nil data", t, funcs.ExcludedMarshal, unmarshal)

	serialization.Set{
		Data: make([]byte, 0),
		Values: []interface{}{
			int8(0), int16(0), int32(0), int64(0), int(0),
			uint8(0), uint16(0), uint32(0), uint64(0), uint(0),
		},
	}.AddModified(mod.All...).Run("unmarshal zero data", t, funcs.ExcludedMarshal, unmarshal)

	serialization.Set{
		Data: []byte("\x00\x00"),
		Values: []interface{}{
			int8(0), int16(0), int32(0), int64(0), int(0),
			uint8(0), uint16(0), uint32(0), uint64(0), uint(0),
		},
	}.AddModified(mod.All...).Run("zeros", t, marshal, unmarshal)

	serialization.Set{
		Data: []byte("\x7f\xff"),
		Values: []interface{}{
			int16(32767), int32(32767), int64(32767), int(32767),
		},
	}.AddModified(mod.All...).Run("32767", t, marshal, unmarshal)

	serialization.Set{
		Data: []byte("\x00\x7f"),
		Values: []interface{}{
			int8(127), int16(127), int32(127), int64(127), int(127),
		},
	}.AddModified(mod.All...).Run("127", t, marshal, unmarshal)

	serialization.Set{
		Data: []byte("\xff\x80"),
		Values: []interface{}{
			int8(-128), int16(-128), int32(-128), int64(-128), int(-128),
		},
	}.AddModified(mod.All...).Run("-128", t, marshal, unmarshal)

	serialization.Set{
		Data: []byte("\x80\x00"),
		Values: []interface{}{
			int16(-32768), int32(-32768), int64(-32768), int(-32768),
		},
	}.AddModified(mod.All...).Run("-32768", t, marshal, unmarshal)

	serialization.Set{
		Data: []byte("\x00\xff"),
		Values: []interface{}{
			uint8(255), uint16(255), uint32(255), uint64(255), uint(255),
		},
	}.AddModified(mod.All...).Run("255", t, marshal, unmarshal)

	serialization.Set{
		Data: []byte("\xff\xff"),
		Values: []interface{}{
			uint16(65535), uint32(65535), uint64(65535), uint(65535),
		},
	}.AddModified(mod.All...).Run("65535", t, marshal, unmarshal)
}
