package help

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/setter"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

var (
	defaultHeaders = []byte("Arguments:\n")

	DefaultWriter = &Writer{
		Factory: setter.DefaultFactory,
		Providers: []provider.Provider{
			env.NewDefaultProvider(),
			flags.NewDefaultProvider(),
		},
	}
)

type Writer struct {
	Providers []provider.Provider
	Factory   setter.Factory
}

func (w *Writer) Write(out io.Writer, in interface{}) (int, error) {
	n, err := out.Write(defaultHeaders)

	if err != nil {
		return n, err
	}

	err = walker.Walk(
		in,
		func(f *walker.Field) error {
			s := w.Factory.Build(f.Field)

			if s == nil {
				return nil
			}

			fks := walker.BuildFieldKeys("", f)

			if len(fks) == 0 {
				return nil
			}

			if f.Value.Type().Implements(setter.ValueType) {
				return walker.SkipStruct
			}

			var b bytes.Buffer

			b.WriteString("\t- ")

			b.WriteString(fks[0])

			b.WriteString(": ")
			b.WriteString(s.String())

			if h, ok := f.Field.Tag.Lookup("help"); ok {
				b.WriteString(" ")
				b.WriteString(h)
			}

			fv := reflectutil.IndirectedValue(f.Value).FieldByName(f.Field.Name)
			if !reflectutil.IsZero(fv) {
				v := reflectutil.IndirectedValue(fv).Interface()

				b.WriteString(" (default: ")

				if ss, ok := v.(fmt.Stringer); ok {
					b.WriteString(ss.String())
				} else {
					fmt.Fprintf(&b, "%+v", v)
				}

				b.WriteString(")")
			}

			var providedKeys []string

			for _, p := range w.Providers {
				if kf, ok := p.(provider.KeyFormatterProvider); ok {
					var ks []string

					for _, k := range walker.BuildFieldKeys(p.StructTag(), f) {
						ks = append(ks, kf.FormatKey(k))
					}

					if len(ks) > 0 {
						providedKeys = append(
							providedKeys,
							fmt.Sprintf("%s: %s", p.StructTag(), strings.Join(ks, ", ")),
						)
					}
				}
			}

			if len(providedKeys) == 0 {
				return nil
			}

			b.WriteString(" (")
			b.WriteString(strings.Join(providedKeys, ", "))
			b.WriteString(")")

			b.WriteRune('\n')

			nn, err := b.WriteTo(out)

			n += int(nn)

			return err
		},
	)

	return n, err
}