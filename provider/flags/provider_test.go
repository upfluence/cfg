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
			name: "has no orphan flags",
			in:   []string{"--no-foo", "--no-bar", "--no-buz"},
			out: map[string]string{
				"foo": "false",
				"bar": "false",
				"buz": "false",
			},
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
		{
			name: "has multiple type format (1)",
			in: []string{
				"--foo",
				"--no-bar",
				"--buz",
				"biz",
				"foo",
				"--foobar",
				"foo",
			},
			out: map[string]string{
				"foo":    "true",
				"bar":    "false",
				"buz":    "biz",
				"foobar": "foo",
			},
		},
		{
			name: "has multiple type format (2)",
			in:   []string{"--foo", "--no-bar", "biz", "foo", "--foobar", "foo"},
			out: map[string]string{
				"foo":    "true",
				"bar":    "false",
				"foobar": "foo",
			},
		},
		{
			name: "has multiple type format (2)",
			in:   []string{"--fuz", "--foo=biz", "--biz=\"buz\""},
			out: map[string]string{
				"foo": "biz",
				"biz": "buz",
				"fuz": "true",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if out := parseFlags(tt.in); !reflect.DeepEqual(out, tt.out) {
				t.Errorf("Wrong flag parsing: %+v instead of %+v", out, tt.out)
			}
		})
	}
}
