package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	for _, tc := range []struct {
		name     string
		haveArgs []string
		want     map[string]string
	}{
		{
			name: "empty",
			want: map[string]string{},
		},
		{
			name:     "no flags",
			haveArgs: []string{"foo", "bar", "buz"},
			want:     map[string]string{},
		},
		{
			name:     "has orphan flags",
			haveArgs: []string{"--foo", "--bar", "--buz"},
			want:     map[string]string{"foo": "true", "bar": "true", "buz": "true"},
		},
		{
			name:     "has no orphan flags",
			haveArgs: []string{"--no-foo", "--no-bar", "--no-buz"},
			want: map[string]string{
				"foo": "false",
				"bar": "false",
				"buz": "false",
			},
		},
		{
			name:     "has multiple values flags",
			haveArgs: []string{"--foo", "bar", "buz"},
			want:     map[string]string{"foo": "bar"},
		},
		{
			name:     "has multiple values single minus flag",
			haveArgs: []string{"-foo", "bar", "buz"},
			want:     map[string]string{"foo": "bar"},
		},
		{
			name: "has multiple type format (1)",
			haveArgs: []string{
				"--foo",
				"--no-bar",
				"--buz",
				"biz",
				"foo",
				"--foobar",
				"foo",
			},
			want: map[string]string{
				"foo":    "true",
				"bar":    "false",
				"buz":    "biz",
				"foobar": "foo",
			},
		},
		{
			name:     "has multiple type format (2)",
			haveArgs: []string{"--foo", "--no-bar", "biz", "foo", "--foobar", "foo"},
			want: map[string]string{
				"foo":    "true",
				"bar":    "false",
				"foobar": "foo",
			},
		},
		{
			name:     "equals syntax",
			haveArgs: []string{"--fuz", "--foo=biz", "--biz=\"buz\""},
			want: map[string]string{
				"foo": "biz",
				"biz": "buz",
				"fuz": "true",
			},
		},
		{
			name:     "equals syntax with embedded equals",
			haveArgs: []string{"--fuz", "--foo=biz", "--biz=\"buz=bar\""},
			want: map[string]string{
				"foo": "biz",
				"biz": "buz=bar",
				"fuz": "true",
			},
		},
		{
			name:     "kebab case flags",
			haveArgs: []string{"--foo-bar", "baz", "--log-level=debug"},
			want: map[string]string{
				"foo-bar":   "baz",
				"log-level": "debug",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, parseFlags(tc.haveArgs))
		})
	}
}

func TestKebabCase(t *testing.T) {
	for _, tc := range []struct {
		name      string
		haveInput string
		want      string
	}{
		{name: "lowercase", haveInput: "foo", want: "foo"},
		{name: "pascal case", haveInput: "FooBar", want: "foo-bar"},
		{name: "camel case", haveInput: "fooBar", want: "foo-bar"},
		{name: "multiple words", haveInput: "FooBarBaz", want: "foo-bar-baz"},
		{name: "single char", haveInput: "F", want: "f"},
		{name: "empty", haveInput: "", want: ""},
		{name: "consecutive uppercase", haveInput: "HTTPServer", want: "http-server"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, kebabCase(tc.haveInput))
		})
	}
}
