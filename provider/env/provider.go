package env

import (
	"context"
	"os"
	"strings"
)

type Provider struct {
	prefix string
}

func NewProvider(p string) *Provider {
	return &Provider{prefix: p}
}

func NewDefaultProvider() *Provider {
	return &Provider{}
}

func (*Provider) StructTag() string { return "env" }

func (p *Provider) buildPrefix() string {
	if p.prefix == "" {
		return ""
	}

	return strings.ToUpper(p.prefix) + "_"
}

func (p *Provider) FormatKey(n string) string {
	return p.buildPrefix() + strings.ToUpper(strings.Replace(n, ".", "_", -1))
}

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	res, ok := os.LookupEnv(p.FormatKey(v))

	return res, ok, nil
}
