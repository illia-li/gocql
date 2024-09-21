package mod

var IntoCustomRef Mod = func(vals ...interface{}) []interface{} {
	return IntoRef(IntoCustom(vals...)...)
}
