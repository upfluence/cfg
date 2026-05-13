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

func (p *Provider) SubKeys(_ context.Context, prefix string) ([]string, error) {
	fullPrefix := p.buildPrefix() + prefix + "_"

	seen := make(map[string]struct{})

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, fullPrefix) {
			continue
		}

		rest := entry[len(fullPrefix):]

		if idx := strings.IndexByte(rest, '='); idx >= 0 {
			rest = rest[:idx]
		}

		if idx := strings.IndexByte(rest, '_'); idx >= 0 {
			rest = rest[:idx]
		}

		if rest == "" {
			continue
		}

		seen[rest] = struct{}{}
	}

	keys := make([]string, 0, len(seen))

	for k := range seen {
		keys = append(keys, k)
	}

	return keys, nil
}
