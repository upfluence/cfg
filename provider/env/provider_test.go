package env

import (
	"context"
	"os"
	"testing"
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
			in:   "foo",
		},
		{
			name:      "simple no prefix",
			in:        "foo",
			envValues: map[string]string{"FOO": "BAR"},
			res:       "BAR",
			exist:     true,
		},
		{
			name:      "pointed value",
			in:        "foo.bar",
			envValues: map[string]string{"FOO_BAR": "BAR"},
			res:       "BAR",
			exist:     true,
		},
		{
			name:      "with prefix",
			in:        "foo",
			prefix:    "pref",
			envValues: map[string]string{"PREF_FOO": "BAR"},
			res:       "BAR",
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
