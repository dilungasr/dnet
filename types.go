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

// ToCtx convets the given value v to *dnet.Ctx.
//
// ok is returned when the assertion succeed
func ToCtx(v any) (ctx *Ctx, ok bool) {
	ctx, ok = v.(*Ctx)
	return ctx, ok
}

// ToEContext convets the given value v to *dnet.EContext.
//
// ok is returned when the assertion succeed
func ToEContext(v any) (ctx *EContext, ok bool) {
	ctx, ok = v.(*EContext)
	return ctx, ok
}
