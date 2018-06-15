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

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	var (
		envVar = p.buildPrefix() + strings.ToUpper(strings.Replace(v, ".", "_", -1))
		res    = os.Getenv(envVar)
	)
	return res, res != "", nil
}
