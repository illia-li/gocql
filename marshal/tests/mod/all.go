package mod

type Mods []Mod

// Mod - value modifiers.
// Designed for test case generators, such as gen.Group, marshal.Group and unmarshal.Group.
type Mod interface {
	Name() string
	Apply([]interface{}) []interface{}
}

var (
	IntoCustom    intoCustom
	IntoRef       intoRef
	IntoCustomRef intoCustomRef

	All = Mods{IntoCustom, IntoRef, IntoCustomRef}
)
