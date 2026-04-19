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

func (*Provider) DefaultFieldValue(fieldName string) string {
	return strings.ToUpper(fieldName)
}

func (*Provider) JoinFieldKeys(prefix, key string) string {
	return prefix + "_" + key
}

func (p *Provider) FormatKey(n string) string {
	return p.buildPrefix() + n
}

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	res, ok := os.LookupEnv(p.buildPrefix() + v)

	return res, ok, nil
}
