package mod

var All = Mods{IntoCustom, IntoRef, IntoCustomRef}

type Mods []Mod

func (l Mods) Append(vals ...interface{}) []interface{} {
	out := append(make([]interface{}, 0), vals...)
	for _, mod := range l {
		out = append(out, mod(vals...)...)
	}
	return out
}

// Mod - value modifiers.
type Mod func(vals ...interface{}) []interface{}

func (m Mod) Append(vals ...interface{}) []interface{} {
	out := append(make([]interface{}, 0), vals...)
	return append(out, m(vals...)...)
}
