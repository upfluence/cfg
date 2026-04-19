package walker

import "reflect"

// Prefixed is an optional interface that a value passed to Walk can
// implement to inject a dynamic ancestor chain.  When Walk receives a
// Prefixed value it uses the ancestor returned by WalkAncestor and
// walks the inner value returned by WalkValue.
type Prefixed interface {
	WalkAncestor() *Field
	WalkValue() any
}

// SubKeyPrefixed is a Prefixed implementation used by the configurator
// and help/synopsis writers to handle dynamic map[string]Struct and
// []Struct fields.  It preserves the real ancestor Field (with struct
// tags) and appends one synthetic segment for the sub-key.
type SubKeyPrefixed struct {
	Ancestor *Field
	SubKey   string
	Value    any
}

func (p *SubKeyPrefixed) WalkAncestor() *Field {
	return &Field{
		Field:    reflect.StructField{Name: p.SubKey},
		Ancestor: p.Ancestor,
	}
}

func (p *SubKeyPrefixed) WalkValue() any { return p.Value }
