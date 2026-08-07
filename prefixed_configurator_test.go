package cfg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/cfg/provider"
)

func TestPrefixed(t *testing.T) {
	type wrapping struct {
		prefix    string
		overrides map[string]string
	}

	for _, tc := range []struct {
		name      string
		have      any
		wrappings []wrapping
		provider  provider.Provider
		want      any
	}{
		{
			name: "default prefix used for provider without override",
			wrappings: []wrapping{
				{prefix: "foo"},
			},
			have: &struct{ Buz int }{},
			provider: &mockProvider{
				st: map[string]string{"foo.Buz": "42"},
			},
			want: &struct{ Buz int }{Buz: 42},
		},
		{
			name: "override replaces prefix segment for matching provider",
			wrappings: []wrapping{
				{prefix: "foo", overrides: map[string]string{"mock": "FOOO"}},
			},
			have: &struct{ Buz int }{},
			provider: &mockProvider{
				st: map[string]string{"FOOO.Buz": "7"},
			},
			want: &struct{ Buz int }{Buz: 7},
		},
		{
			name: "non-overridden provider falls back to raw prefix",
			wrappings: []wrapping{
				{prefix: "foo", overrides: map[string]string{"other": "X"}},
			},
			have: &struct{ Buz int }{},
			provider: &mockProvider{
				st: map[string]string{"foo.Buz": "99"},
			},
			want: &struct{ Buz int }{Buz: 99},
		},
		{
			name: "chained Prefixed stacks ancestors",
			wrappings: []wrapping{
				{prefix: "a"},
				{prefix: "b"},
			},
			have: &struct{ Val string }{},
			provider: &mockProvider{
				st: map[string]string{"a.b.Val": "nested"},
			},
			want: &struct{ Val string }{Val: "nested"},
		},
		{
			name: "nested struct under prefix",
			wrappings: []wrapping{
				{prefix: "pfx"},
			},
			have: &struct {
				Inner struct {
					V int `mock:"v"`
				} `mock:"inner"`
			}{},
			provider: &mockProvider{
				st: map[string]string{"pfx.inner.v": "5"},
			},
			want: func() any {
				v := &struct {
					Inner struct {
						V int `mock:"v"`
					} `mock:"inner"`
				}{}

				v.Inner.V = 5

				return v
			}(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pc := Configurator(NewConfiguratorWithOptions(WithProviders(tc.provider)))

			for _, w := range tc.wrappings {
				pc = pc.Prefixed(w.prefix, w.overrides)
			}

			err := pc.Populate(context.Background(), tc.have)

			require.NoError(t, err)
			assert.Equal(t, tc.want, tc.have)
		})
	}
}
