package synopsis

import (
	"bytes"
	"io"
	"reflect"

	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/setter"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/flags"
)

var DefaultWriter = &Writer{
	Factory:  setter.DefaultFactory,
	Provider: flags.NewProvider(nil),
}

type Writer struct {
	Factory          setter.Factory
	Provider         provider.Provider
	IgnoreMissingTag bool
}

func (w *Writer) Write(out io.Writer, in interface{}) (int, error) {
	var b bytes.Buffer

	writeFn := w.buildWriteFn(&b)

	if err := walker.Walk(in, writeFn); err != nil {
		return 0, err
	}

	return out.Write(b.Bytes())
}

func (w *Writer) buildWriteFn(b *bytes.Buffer) walker.WalkFunc {
	return func(f *walker.Field) error {
		if s := w.Factory.Build(f.Field.Type); s == nil {
			return w.writeSubKeyField(b, f)
		}

		fks := walker.BuildFieldKeys(
			provider.WrapFullyQualifiedProvider(w.Provider),
			f,
			w.IgnoreMissingTag,
		)

		if len(fks) == 0 {
			return nil
		}

		if setter.IsUnmarshaler(f.Value.Type()) {
			return walker.SkipStruct
		}

		w.writeKeys(b, fks)

		return nil
	}
}

func (w *Writer) writeKeys(b *bytes.Buffer, fks []string) {
	b.WriteRune('[')

	kf, hasFormatter := w.Provider.(provider.KeyFormatter)

	for i, fk := range fks {
		if hasFormatter {
			fk = kf.FormatKey(fk)
		}

		b.WriteString(fk)

		if i < len(fks)-1 {
			b.WriteString(", ")
		}
	}

	b.WriteString("] ")
}

func (w *Writer) writeSubKeyField(b *bytes.Buffer, f *walker.Field) error {
	var (
		placeholder string
		structType  reflect.Type
	)

	if st := reflectutil.SubKeyMapElem(f.Field.Type); st != nil {
		placeholder = "<key>"
		structType = st
	} else if st := reflectutil.SubKeySliceElem(f.Field.Type); st != nil {
		placeholder = "<N>"
		structType = st
	} else {
		return nil
	}

	prefix := f.FieldPrefix()
	prefix = append(prefix, placeholder)
	elem := reflect.New(structType)

	prefixed := &walker.SubKeyPrefixed{Prefix: prefix, Value: elem.Interface()}

	return walker.Walk(prefixed, w.buildWriteFn(b))
}
