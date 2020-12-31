package synopsis

import (
	"bytes"
	"io"

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
	Factory  setter.Factory
	Provider provider.KeyFormatterProvider
}

func (w *Writer) Write(out io.Writer, in interface{}) (int, error) {
	var b bytes.Buffer

	if err := walker.Walk(
		in,
		func(f *walker.Field) error {
			if s := w.Factory.Build(f.Field); s == nil {
				return nil
			}

			fks := walker.BuildFieldKeys(w.Provider.StructTag(), f)

			if len(fks) == 0 {
				return nil
			}

			if f.Value.Type().Implements(setter.ValueType) {
				return walker.SkipStruct
			}

			b.WriteRune('[')

			for i, fk := range fks {
				b.WriteString(w.Provider.FormatKey(fk))

				if i < len(fks)-1 {
					b.WriteString(", ")
				}
			}

			b.WriteString("] ")

			return nil
		},
	); err != nil {
		return 0, err
	}

	return out.Write(b.Bytes())
}
