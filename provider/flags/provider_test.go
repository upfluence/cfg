package flags

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   []string
		out  map[string]string
	}{
		{
			name: "empty",
			out:  map[string]string{},
		},
		{
			name: "no flags",
			in:   []string{"foo", "bar", "buz"},
			out:  map[string]string{},
		},
		{
			name: "has orphan flags",
			in:   []string{"--foo", "--bar", "--buz"},
			out:  map[string]string{"foo": "true", "bar": "true", "buz": "true"},
		},
		{
			name: "has multiple values flags",
			in:   []string{"--foo", "bar", "buz"},
			out:  map[string]string{"foo": "bar"},
		},
		{
			name: "has multiple values single minus flag",
			in:   []string{"-foo", "bar", "buz"},
			out:  map[string]string{"foo": "bar"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if out := parseFlags(tt.in); !reflect.DeepEqual(out, tt.out) {
				t.Errorf("Wrong flag parsing: %+v instead of %+v", out, tt.out)
			}
		})
	}
}
