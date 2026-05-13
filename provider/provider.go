package provider

import "context"

type Provider interface {
	StructTag() string
	Provide(context.Context, string) (string, bool, error)
}

func ProvideError(tag string, err error) Provider {
	return &faultyProvider{tag: tag, err: err}
}

type faultyProvider struct {
	tag string
	err error
}

func (p *faultyProvider) Err() error        { return p.err }
func (p *faultyProvider) StructTag() string { return p.tag }
func (p *faultyProvider) Provide(context.Context, string) (string, bool, error) {
	return "", false, p.err
}

type KeyFn func(string) string

func NewStaticProvider(tag string, vs map[string]string, kfn KeyFn) Provider {
	if kfn == nil {
		kfn = func(k string) string { return k }
	}

	return &staticProvider{vs: vs, tag: tag, kfn: kfn}
}

type staticProvider struct {
	vs  map[string]string
	tag string
	kfn KeyFn
}

func (sp *staticProvider) StructTag() string { return sp.tag }

func (sp *staticProvider) Provide(_ context.Context, k string) (string, bool, error) {
	v, ok := sp.vs[sp.kfn(k)]

	return v, ok, nil
}

// FullyQualifiedProvider is an optional interface that providers can
// implement to customise how struct field keys are built by the walker.
//
// DefaultFieldValue returns the key to use when a struct field does not
// carry the provider's struct tag.  Returning "" causes the field to be
// skipped for this provider.
//
// JoinFieldKeys controls how ancestor and leaf keys are concatenated.
// The standard provider joins them with "."; other providers may need a
// different strategy (e.g. the default-value provider filters out empty
// parts).
//
// SubKeys enumerates dynamic sub-keys under a given prefix.  This is
// used by the configurator to populate map[string]Struct fields: for
// each discovered sub-key a new struct value is allocated and
// recursively walked with the sub-key injected as an ancestor prefix.
// Providers that do not support key enumeration should return a nil
// slice.
type FullyQualifiedProvider interface {
	Provider

	DefaultFieldValue(fieldName string) string
	JoinFieldKeys(prefix, key string) string
	SubKeys(ctx context.Context, prefix string) ([]string, error)
}

// WrapFullyQualifiedProvider returns p as a FullyQualifiedProvider.  If
// p already implements the interface it is returned as-is; otherwise it
// is wrapped with standard defaults (field name fallback and dot-joined
// keys).
func WrapFullyQualifiedProvider(p Provider) FullyQualifiedProvider {
	if fqp, ok := p.(FullyQualifiedProvider); ok {
		return fqp
	}

	return &defaultFQProvider{Provider: p}
}

type defaultFQProvider struct {
	Provider
}

func (d *defaultFQProvider) DefaultFieldValue(fieldName string) string {
	return fieldName
}

func (d *defaultFQProvider) JoinFieldKeys(prefix, key string) string {
	return prefix + "." + key
}

func (*defaultFQProvider) SubKeys(context.Context, string) ([]string, error) {
	return nil, nil
}

// KeyFormatter is an optional interface that providers can implement to
// control how keys are displayed in help and synopsis output.
type KeyFormatter interface {
	FormatKey(string) string
}
