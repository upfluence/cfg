package flags

import (
	"context"
	"os"
	"strings"

	"github.com/upfluence/cfg/provider"
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
	)

	for _, arg := range args {

		if v, ok := parseArg(arg); ok {

			if strings.HasPrefix(v, "no-") && (len(v) > 3) {
				key = strings.Replace(v, "no-", "", 1)
				res[key] = "false"
			} else {
				key = v
				res[key] = "true"
			}
			inParam = true
		} else if (len(v) > 0) && inParam {

			res[key] = v
			inParam = false
		}
	}

	return res
}

func NewDefaultProvider() provider.Provider {
	return NewProvider(os.Args[1:])
}

func NewProvider(args []string) provider.Provider {
	return &Provider{store: parseFlags(args)}
}

func (*Provider) StructTag() string { return "flag" }

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	var res, ok = p.store[v]

	return res, ok, nil
}
