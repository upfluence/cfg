package cfg

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/json"
)

type config struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
}

func jsonProvider(v string) provider.Provider {
	return json.NewProviderFromReader(strings.NewReader(v))
}

func TestIntegration(t *testing.T) {
	for _, tt := range []struct {
		stages []Stage

		want config
	}{
		{want: config{Foo: "bar"}},
		{
			stages: []Stage{
				ConfigurationStage[config]{
					NextProvidersFunc: func(c config) []provider.Provider {
						return []provider.Provider{
							jsonProvider(`{"bar":"buz"}`),
						}
					},
				},
			},
			want: config{Foo: "bar", Bar: "buz"},
		},
		{
			stages: []Stage{
				ConfigurationStage[config]{
					Mode: ProviderReplace,
					NextProvidersFunc: func(c config) []provider.Provider {
						return []provider.Provider{jsonProvider(`{"bar":"buz"}`)}
					},
				},
			},
			want: config{Bar: "buz"},
		},
	} {
		c := Configurator{
			Stages: tt.stages,
			InitialConfigurator: cfg.NewConfigurator(
				jsonProvider(`{"foo":"bar"}`),
			),
		}

		var v config

		err := c.Populate(context.Background(), &v)

		require.NoError(t, err)

		assert.Equal(t, tt.want, v)
	}
}
