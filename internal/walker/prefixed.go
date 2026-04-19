package walker

import (
	"reflect"

	"github.com/upfluence/cfg/internal/reflectutil"
)

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

// BuildSubKeyField returns a SubKeyPrefixed for a map[string]Struct or
// []Struct field, using "<key>" or "<N>" as the placeholder sub-key.
// It returns nil if the field type is neither.
func BuildSubKeyField(f *Field) *SubKeyPrefixed {
	var (
		placeholder string
		structType  reflect.Type
	)

	if st := reflectutil.SubKeyMapElem(f.Field.Type); st != nil {
		placeholder = "<key>"
		structType = st
	} else if st := reflectutil.SubKeySliceElem(f.Field.Type); st != nil {
		placeholder = "<N>"
		structType = st
	} else {
		return nil
	}

	return &SubKeyPrefixed{
		Ancestor: f,
		SubKey:   placeholder,
		Value:    reflect.New(structType).Interface(),
	}
}
