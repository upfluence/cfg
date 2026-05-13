package env

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_Provide(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		in        string
		res       string
		exist     bool
		wantErr   bool
		envValues map[string]string
	}{
		{
			name: "empty",
			in:   "FOO",
		},
		{
			name:      "simple no prefix",
			in:        "FOO",
			envValues: map[string]string{"FOO": "BAR"},
			res:       "BAR",
			exist:     true,
		},
		{
			name:      "nested value",
			in:        "FOO_BAR",
			envValues: map[string]string{"FOO_BAR": "BAR"},
			res:       "BAR",
			exist:     true,
		},
		{
			name:      "with prefix",
			in:        "FOO",
			prefix:    "pref",
			envValues: map[string]string{"PREF_FOO": "BAR"},
			res:       "BAR",
			exist:     true,
		},
		{
			name:      "empty string is present",
			in:        "FOO",
			envValues: map[string]string{"FOO": ""},
			res:       "",
			exist:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envValues {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			p := &Provider{prefix: tt.prefix}

			got, got1, err := p.Provide(context.Background(), tt.in)

			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Provide() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.res {
				t.Errorf("Provider.Provide() got = %v, want %v", got, tt.res)
			}

			if got1 != tt.exist {
				t.Errorf("Provider.Provide() got1 = %v, want %v", got1, tt.exist)
			}
		})
	}
}

func TestProvider_SubKeys(t *testing.T) {
	for _, tc := range []struct {
		name    string
		haveEnv map[string]string
		haveKey string
		prefix  string
		want    []string
	}{
		{
			name:    "no matching vars",
			haveKey: "DB",
			want:    []string{},
		},
		{
			name:    "single sub-key",
			haveEnv: map[string]string{"DB_PRIMARY_HOST": "localhost"},
			haveKey: "DB",
			want:    []string{"PRIMARY"},
		},
		{
			name: "multiple sub-keys",
			haveEnv: map[string]string{
				"DB_PRIMARY_HOST":   "h1",
				"DB_PRIMARY_PORT":   "5432",
				"DB_REPLICA_HOST":   "h2",
				"DB_SECONDARY_HOST": "h3",
			},
			haveKey: "DB",
			want:    []string{"PRIMARY", "REPLICA", "SECONDARY"},
		},
		{
			name: "with global prefix",
			haveEnv: map[string]string{
				"APP_CACHE_REDIS_HOST":    "r1",
				"APP_CACHE_MEMCACHE_HOST": "m1",
				"CACHE_OTHER_HOST":        "x",
			},
			haveKey: "CACHE",
			prefix:  "app",
			want:    []string{"MEMCACHE", "REDIS"},
		},
		{
			name: "ignores vars with empty segment after prefix",
			haveEnv: map[string]string{
				"DB_HOST": "localhost",
			},
			haveKey: "DB_HOST",
			want:    []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.haveEnv {
				t.Setenv(k, v)
			}

			p := &Provider{prefix: tc.prefix}

			got, err := p.SubKeys(context.Background(), tc.haveKey)

			require.NoError(t, err)

			sort.Strings(got)
			sort.Strings(tc.want)
			assert.Equal(t, tc.want, got)
		})
	}
}
