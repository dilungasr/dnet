package dnet

// typ brings together the utilities for working with dnet types
type typ struct{}

var Type typ = typ{}

// IsEContext tests the given value v if it's an external context - EContext
func (t typ) IsEContext(v any) bool {
	_, ok := v.(EContext)
	return ok
}

// IsCtx tests the given value v if it's of Ctx type
func (t typ) IsCtxt(v any) bool {
	_, ok := v.(Ctx)
	return ok
}
