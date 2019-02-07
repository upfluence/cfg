package json

import (
	"context"
	"testing"
)

func TestProvider_Provide(t *testing.T) {
	tests := []struct {
		name        string
		p           *Provider
		in          string
		wantValue   string
		wantValueFn func(string) bool
		wantExist   bool
		wantErr     bool
	}{
		{
			name: "empty store",
			p:    &Provider{},
		},
		{
			name:      "top level value",
			p:         &Provider{map[string]interface{}{"foo": "bar"}},
			in:        "foo",
			wantValue: "bar",
			wantExist: true,
		},
		{
			name:      "slice value",
			p:         &Provider{map[string]interface{}{"foo": []int64{1, 2, 3}}},
			in:        "foo",
			wantValue: "1,2,3",
			wantExist: true,
		},
		{
			name: "map value",
			p: &Provider{
				map[string]interface{}{
					"foo": map[string]int64{"foo": 1, "bar": 2},
				},
			},
			in: "foo",
			wantValueFn: func(got string) bool {
				switch got {
				case "foo=1,bar=2", "bar=2,foo=1":
					return true
				}

				return false
			},
			wantExist: true,
		},
		{
			name:      "second level value",
			p:         &Provider{map[string]interface{}{"foo": map[string]interface{}{"fiz": "bar"}}},
			in:        "foo.fiz",
			wantValue: "bar",
			wantExist: true,
		},
		{
			name:    "wrong format",
			p:       &Provider{map[string]interface{}{"foo": map[string]interface{}{"fiz": "bar"}}},
			in:      "foo.fiz.buz",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.p.Provide(context.Background(), tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantValueFn != nil {
				if !tt.wantValueFn(got) {
					t.Errorf("Provider.Provide() got = %v", got)
				}
			} else if got != tt.wantValue {
				t.Errorf("Provider.Provide() got = %v, want %v", got, tt.wantValue)
			}
			if got1 != tt.wantExist {
				t.Errorf("Provider.Provide() got1 = %v, want %v", got1, tt.wantExist)
			}
		})
	}
}
