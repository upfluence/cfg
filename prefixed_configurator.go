package cfg

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/upfluence/cfg/internal/walker"
)

// prefixedValue implements walker.Prefixed to inject a synthetic prefix
// ancestor when the user struct is walked.
type prefixedValue struct {
	ancestor *walker.Field
	value    any
}

func (pv *prefixedValue) WalkAncestor() *walker.Field { return pv.ancestor }
func (pv *prefixedValue) WalkValue() any              { return pv.value }

// buildPrefixField constructs a synthetic *walker.Field whose struct tag
// encodes each override entry and whose Name is the raw prefix (used as
// the default key segment by providers that have no override).
func buildPrefixField(prefix string, overrides map[string]string, ancestor *walker.Field) *walker.Field {
	var parts []string

	for tag, seg := range overrides {
		parts = append(parts, tag+`:"`+seg+`"`)
	}

	// Sort for deterministic tag string ordering.
	sort.Strings(parts)

	return &walker.Field{
		Field: reflect.StructField{
			Name: prefix,
			Tag:  reflect.StructTag(strings.Join(parts, " ")),
		},
		Ancestor: ancestor,
	}
}

type prefixedConfigurator struct {
	inner    *configurator
	ancestor *walker.Field
}

func (pc *prefixedConfigurator) Populate(ctx context.Context, in any) error {
	return pc.inner.Populate(ctx, &prefixedValue{ancestor: pc.ancestor, value: in})
}

func (pc *prefixedConfigurator) WithOptions(opts ...Option) Configurator {
	return &prefixedConfigurator{
		inner:    pc.inner.withOptions(opts),
		ancestor: pc.ancestor,
	}
}

func (pc *prefixedConfigurator) Prefixed(prefix string, overrides map[string]string) Configurator {
	return &prefixedConfigurator{
		inner:    pc.inner,
		ancestor: buildPrefixField(prefix, overrides, pc.ancestor),
	}
}
