package flags

import (
	"context"
	"os"
)

type Provider struct {
	store map[string]string
}

func parseArg(s string) (string, bool) {
	if len(s) < 2 || s[0] != '-' {
		return s, false
	}

	numMinuses := 1

	if s[1] == '-' {
		numMinuses++
	}

	return s[numMinuses:], (len(s) - numMinuses) > 0
}

func parseFlags(args []string) map[string]string {
	var (
		res = make(map[string]string)

		key     string
		inParam bool

		pushSingleKey = func() {
			res[key] = "true"
			inParam = false
		}
	)

	for _, arg := range args {
		v, ok := parseArg(arg)

		if ok {
			if inParam {
				pushSingleKey()
			}

			key = v
			inParam = true
		} else if len(v) > 0 && inParam {
			res[key] = v
			inParam = false
		}
	}

	if inParam {
		pushSingleKey()
	}

	return res
}

func NewDefaultProvider() *Provider {
	return NewProvider(os.Args[1:])
}

func NewProvider(args []string) *Provider {
	return &Provider{store: parseFlags(args)}
}

func (*Provider) StructTag() string { return "flag" }

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	var res, ok = p.store[v]

	return res, ok, nil
}
