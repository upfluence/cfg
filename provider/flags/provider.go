package flags

import (
	"context"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/upfluence/cfg/internal/stringutil"
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

			switch vs, _ := stringutil.Split(v, '='); len(vs) {
			case 2:
				if inParam {
					res[key] = val
				}

				inParam = false
				key = vs[0]
				val = vs[1]

				if v, err := strconv.Unquote(val); err == nil {
					val = v
				}
			case 1:
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

func NewDefaultProvider() *Provider {
	return NewProvider(os.Args[1:])
}

func NewProvider(args []string) *Provider {
	return &Provider{
		sp: provider.NewStaticProvider(
			StructTag,
			parseFlags(args),
			strings.ToLower,
		),
	}
}

type Provider struct {
	sp provider.Provider
}

func kebabCase(s string) string {
	var b strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				if unicode.IsLower(rune(s[i-1])) || (len(s) > i+1 && unicode.IsLower(rune(s[i+1]))) {
					b.WriteByte('-')
				}
			}

			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

func (*Provider) StructTag() string { return StructTag }

func (*Provider) DefaultFieldValue(fieldName string) string {
	return kebabCase(fieldName)
}

func (*Provider) JoinFieldKeys(prefix, key string) string {
	return prefix + "." + key
}

func (*Provider) FormatKey(n string) string {
	n = strings.ToLower(n)

	if len(n) == 1 {
		return "-" + n
	}

	return "--" + n
}

func (p *Provider) Provide(ctx context.Context, k string) (string, bool, error) {
	return p.sp.Provide(ctx, k)
}
