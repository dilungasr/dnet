package dnet

// IsEContext tests the given value v if it's an external context - EContext
func IsEContext(v any) bool {
	_, ok1 := v.(EContext)
	_, ok2 := v.(*EContext)
	return ok1 || ok2
}

// IsCtx tests the given value v if it's of Ctx type
func IsCtxt(v any) bool {
	_, ok1 := v.(Ctx)
	_, ok2 := v.(*Ctx)
	return ok1 || ok2
}

// IsDnetContext tests the given  value v if it's  a dnet context i.e dnet.EContext or dnet.Ctx
func IsDnetContext(v any) bool {
	return IsCtxt(v) || IsEContext(v)
}
