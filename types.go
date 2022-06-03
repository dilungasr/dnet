package dnet

type DnetContext interface {
	*Ctx | *EContext
}

// IsEContext tests the given value v if it's an external context - EContext
func IsEContext(v any) bool {
	_, ok := v.(*EContext)
	return ok
}

// IsCtx tests the given value v if it's of Ctx type
func IsCtxt(v any) bool {
	_, ok := v.(*Ctx)
	return ok
}

// IsDnetContext tests the given  value v if it's  a dnet context i.e dnet.EContext or dnet.Ctx
func IsDnetContext(v any) bool {
	return IsCtxt(v) || IsEContext(v)
}

// ToCtx converts the given value v to *dnet.Ctx.
//
// ok is returned when the assertion succeed
func ToCtx(v any) (ctx *Ctx, ok bool) {
	ctx, ok = v.(*Ctx)
	return ctx, ok
}

// ToEContext converts the given value v to *dnet.EContext.
//
// ok is returned when the assertion succeed
func ToEContext(v any) (ctx *EContext, ok bool) {
	ctx, ok = v.(*EContext)
	return ctx, ok
}
