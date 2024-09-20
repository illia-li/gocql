package gocql

import (
	"testing"

	"github.com/gocql/gocql/marshal/tests/mod"
	"github.com/gocql/gocql/marshal/tests/mustfail/marshal"
	"github.com/gocql/gocql/marshal/tests/mustfail/unmarshal"
)

func TestMarshalSmallintMustFail(t *testing.T) {
	tFunc := func(i interface{}) ([]byte, error) { return Marshal(NativeType{proto: 4, typ: TypeTinyInt}, i) }

	marshal.Set{
		MarshalIns: []interface{}{
			int32(32768), int64(32768), int(32768), "32768",
			int32(-32769), int64(-32769), int(-32769), "-32769",
			uint32(65536), uint64(65536), uint(65536),
		},
	}.AddModified(mod.All...).Run("big_vals", t, tFunc)

	marshal.Set{
		MarshalIns: []interface{}{"1s2", "1s", "-1s", ".1", ",1", "0.1", "0,1"},
	}.AddModified(mod.All...).Run("corrupt_vals", t, tFunc)
}

func TestUnmarshalSmallintMustFail(t *testing.T) {
	tType := NativeType{proto: 4, typ: TypeTinyInt}
	tFunc := func(d []byte, i interface{}) error { return Unmarshal(tType, d, i) }

	unmarshal.Set{
		Data: []byte("\x80\x00\x00"),
		UnmarshalIns: []interface{}{
			int8(0), int16(0), int32(0), int64(0), int(0), "",
			uint8(0), uint16(0), uint32(0), uint64(0), uint(0),
		},
		Issue: "https://github.com/scylladb/gocql/issues/246",
	}.AddModified(mod.All...).Run("big_data", t, tFunc)

	unmarshal.Set{
		Data: []byte("\x80"),
		UnmarshalIns: []interface{}{
			int8(0), int16(0), int32(0), int64(0), int(0), "",
			uint8(0), uint16(0), uint32(0), uint64(0), uint(0),
		},
		Issue: "https://github.com/scylladb/gocql/issues/252",
	}.AddModified(mod.All...).Run("small_data", t, tFunc)

	unmarshal.Set{
		Data:         []byte("\x7f\xff"),
		UnmarshalIns: []interface{}{int8(0)},
		Issue:        "https://github.com/scylladb/gocql/issues/253",
	}.AddModified(mod.All...).Run("small_val_type_+int", t, tFunc)

	unmarshal.Set{
		Data:         []byte("\x80\x00"),
		UnmarshalIns: []interface{}{int8(0)},
		Issue:        "https://github.com/scylladb/gocql/issues/253",
	}.AddModified(mod.All...).Run("small_val_type_-int", t, tFunc)

	unmarshal.Set{
		Data:         []byte("\xff\xff"),
		UnmarshalIns: []interface{}{uint8(0)},
		Issue:        "https://github.com/scylladb/gocql/issues/253",
	}.AddModified(mod.All...).Run("small_val_type_uint", t, tFunc)

}
