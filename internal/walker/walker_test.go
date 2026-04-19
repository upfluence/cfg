package walker

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors/errtest"
)

type buz struct {
	Foo string
}

type baz struct {
	Struct    buz
	StructPtr *buz
}

func dotPath(f *Field) string {
	a := f.Ancestor
	vs := []string{f.Field.Name}

	for a != nil {
		vs = append(vs, a.Field.Name)
		a = a.Ancestor
	}

	return strings.Join(vs, ".")
}

type foo struct {
	buz
}

type prefixed struct {
	prefix []string
	value  any
}

func (p *prefixed) WalkAncestor() *Field {
	var ancestor *Field

	for _, seg := range p.prefix {
		ancestor = &Field{
			Field:    reflect.StructField{Name: seg},
			Ancestor: ancestor,
		}
	}

	return ancestor
}

func (p *prefixed) WalkValue() any { return p.value }

type outerWithPrefixed struct {
	Nested *prefixed
}

func TestWalk(t *testing.T) {
	var (
		castedBazNil *baz
		castedIntNil *int
	)

	for _, tt := range []struct {
		name string
		in   interface{}

		outfn func(*testing.T, []string)
		errfn errtest.ErrorAssertion
	}{
		{
			name:  "nil",
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: errtest.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name:  "casted nil",
			in:    castedBazNil,
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: errtest.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name:  "casted int",
			in:    castedIntNil,
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: errtest.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name:  "baz",
			in:    baz{},
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: errtest.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name: "baz ptr",
			in:   &baz{},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(
					t,
					[]string{"Struct", "Foo.Struct", "StructPtr", "Foo.StructPtr"},
					vs,
				)
			},
			errfn: errtest.NoError(),
		},
		{
			name: "foo ptr",
			in:   &foo{},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(t, []string{"Foo.buz"}, vs)
			},
			errfn: errtest.NoError(),
		},
		{
			name: "prefixed nil value",
			in:   &prefixed{prefix: []string{"x"}, value: nil},
			outfn: func(t *testing.T, vs []string) {
				assert.Empty(t, vs)
			},
			errfn: errtest.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name: "prefixed single segment",
			in:   &prefixed{prefix: []string{"pfx"}, value: &buz{}},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(t, []string{"Foo.pfx"}, vs)
			},
			errfn: errtest.NoError(),
		},
		{
			name: "prefixed multiple segments",
			in:   &prefixed{prefix: []string{"a", "b"}, value: &buz{}},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(t, []string{"Foo.b.a"}, vs)
			},
			errfn: errtest.NoError(),
		},
		{
			name: "prefixed nested struct",
			in:   &prefixed{prefix: []string{"ns"}, value: &baz{}},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(
					t,
					[]string{"Struct.ns", "Foo.Struct.ns", "StructPtr.ns", "Foo.StructPtr.ns"},
					vs,
				)
			},
			errfn: errtest.NoError(),
		},
		{
			name: "prefixed empty prefix",
			in:   &prefixed{prefix: nil, value: &buz{}},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(t, []string{"Foo"}, vs)
			},
			errfn: errtest.NoError(),
		},
		{
			name: "nested prefixed field",
			in: &outerWithPrefixed{
				Nested: &prefixed{prefix: []string{"dyn"}, value: &buz{}},
			},
			outfn: func(t *testing.T, vs []string) {
				assert.Equal(t, []string{"Nested", "Foo.dyn.Nested"}, vs)
			},
			errfn: errtest.NoError(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var vs []string

			err := Walk(tt.in, func(f *Field) error {
				vs = append(vs, dotPath(f))

				return nil
			})

			tt.outfn(t, vs)
			tt.errfn.Assert(t, err)
		})
	}
}
