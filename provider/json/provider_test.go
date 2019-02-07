package json

import (
	"context"
	"testing"
)

func TestProvider_Provide(t *testing.T) {
	tests := []struct {
		name      string
		p         *Provider
		in        string
		wantValue string
		wantExist bool
		wantErr   bool
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
			if got != tt.wantValue {
				t.Errorf("Provider.Provide() got = %v, want %v", got, tt.wantValue)
			}
			if got1 != tt.wantExist {
				t.Errorf("Provider.Provide() got1 = %v, want %v", got1, tt.wantExist)
			}
		})
	}
}
