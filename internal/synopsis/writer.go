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
	Factory          setter.Factory
	Provider         provider.Provider
	IgnoreMissingTag bool
}

func (w *Writer) Write(out io.Writer, in interface{}) (int, error) {
	var b bytes.Buffer

	if err := walker.Walk(
		in,
		func(f *walker.Field) error {
			if s := w.Factory.Build(f.Field.Type); s == nil {
				return nil
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

			return nil
		},
	); err != nil {
		return 0, err
	}

	return out.Write(b.Bytes())
}
