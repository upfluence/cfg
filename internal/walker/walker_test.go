package walker

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/testutil"
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

func TestWalk(t *testing.T) {
	var (
		castedBazNil *baz
		castedIntNil *int
	)

	for _, tt := range []struct {
		name string
		in   interface{}

		outfn func(*testing.T, []string)
		errfn testutil.ErrorAssertion
	}{
		{
			name:  "nil",
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: testutil.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name:  "casted nil",
			in:    castedBazNil,
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: testutil.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name:  "casted int",
			in:    castedIntNil,
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: testutil.ErrorEqual(ErrShouldBeAStructPtr),
		},
		{
			name:  "baz",
			in:    baz{},
			outfn: func(t *testing.T, vs []string) { assert.Equal(t, 0, len(vs)) },
			errfn: testutil.ErrorEqual(ErrShouldBeAStructPtr),
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
			errfn: testutil.NoError(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var vs []string

			err := Walk(tt.in, func(f *Field) error {
				vs = append(vs, dotPath(f))

				return nil
			})

			tt.outfn(t, vs)
			tt.errfn(t, err)
		})
	}
}
