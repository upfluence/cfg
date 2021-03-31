package flags

import (
	"os"
	"strconv"
	"strings"

	"github.com/upfluence/cfg/provider"
)

const StructTag = "flag"

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
			val := "true"

			if strings.Contains(v, "=") {
				if inParam {
					res[key] = val
				}

				vs := strings.SplitN(v, "=", 2)
				inParam = false
				key = vs[0]
				val = vs[1]

				if v, err := strconv.Unquote(val); err == nil {
					val = v
				}
			} else {
				key = v
				inParam = true

				if strings.HasPrefix(v, "no-") && len(v) > 3 {
					key = strings.TrimPrefix(v, "no-")
					val = "false"
					inParam = false
				}
			}

			res[key] = val
		} else if len(v) > 0 && inParam {
			res[key] = v
			inParam = false
		}
	}

	return res
}

func NewDefaultProvider() provider.KeyFormatterProvider {
	return NewProvider(os.Args[1:])
}

func NewProvider(args []string) provider.KeyFormatterProvider {
	return provider.KeyFormatterFunc{
		Provider: provider.NewStaticProvider(
			StructTag,
			parseFlags(args),
			strings.ToLower,
		),
		KeyFormatFunc: func(n string) string {
			n = strings.ToLower(n)

			if len(n) == 1 {
				return "-" + n
			}

			return "--" + n
		},
	}
}
