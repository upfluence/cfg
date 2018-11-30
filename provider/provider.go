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

func (p *faultyProvider) StructTag() string { return p.tag }
func (p *faultyProvider) Provide(context.Context, string) (string, bool, error) {
	return "", false, p.err
}
