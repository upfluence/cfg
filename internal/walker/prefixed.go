package walker

// Prefixed is an optional interface that a value passed to Walk can
// implement to inject dynamic key prefix segments.  When Walk receives a
// Prefixed value it builds a synthetic ancestor chain from the prefix
// segments and walks the inner value returned by WalkValue.
type Prefixed interface {
	WalkPrefix() []string
	WalkValue() any
}

// SubKeyPrefixed is a Prefixed implementation that injects a list of
// prefix segments before walking an inner struct value.  It is used by
// the configurator and help/synopsis writers to handle dynamic
// map[string]Struct and []Struct fields.
type SubKeyPrefixed struct {
	Prefix []string
	Value  any
}

func (p *SubKeyPrefixed) WalkPrefix() []string { return p.Prefix }
func (p *SubKeyPrefixed) WalkValue() any       { return p.Value }
