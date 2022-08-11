package providertest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors/errtest"

	"github.com/upfluence/cfg/provider"
)

type ProvideAssertion func(testing.TB, provider.Provider, string)

func AssertError(ea errtest.ErrorAssertion) func(testing.TB, provider.Provider, string) {
	return func(t testing.TB, p provider.Provider, k string) {
		v, ok, err := p.Provide(context.Background(), k)

		assert.Equal(t, "", v)
		assert.False(t, ok)
		ea.Assert(t, err)
	}
}

type ValueFunc func(string) (string, bool)

func StaticValue(k, v string) func(string) (string, bool) {
	return func(kk string) (string, bool) {
		if k == kk {
			return v, true
		}

		return "", false
	}
}

func MapValue(vs map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := vs[k]

		return v, ok
	}
}

func AssertValues(vfn ValueFunc) func(testing.TB, provider.Provider, string) {
	return func(t testing.TB, p provider.Provider, k string) {
		v, ok, err := p.Provide(context.Background(), k)

		assert.Nil(t, err)

		tv, tok := vfn(k)
		assert.Equal(t, tv, v)
		assert.Equal(t, tok, ok)
	}
}
