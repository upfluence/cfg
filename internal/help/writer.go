package help

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/setter"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	dflt "github.com/upfluence/cfg/provider/default"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

var (
	defaultHeaders = []byte("Arguments:\n")

	DefaultWriter = &Writer{
		Factory: setter.DefaultFactory,
		Providers: []provider.Provider{
			dflt.Provider{},
			env.NewDefaultProvider(),
			flags.NewDefaultProvider(),
		},
	}

	helperType = reflect.TypeOf((*helper)(nil)).Elem()
)

type helper interface {
	Help() string
}

func fieldHelp(f *walker.Field) string {
	fv := reflectutil.IndirectedValue(f.Value).FieldByName(f.Field.Name)

	if fv.CanAddr() && fv.Addr().Type().Implements(helperType) {
		if h := fv.Addr().Interface().(helper).Help(); h != "" {
			return h
		}
	} else if fv.Type().Implements(helperType) && fv.CanInterface() {
		if h := fv.Interface().(helper).Help(); h != "" {
			return h
		}
	}

	if h, ok := f.Field.Tag.Lookup("help"); ok {
		return h
	}

	return ""
}

type Writer struct {
	Providers        []provider.Provider
	Factory          setter.Factory
	IgnoreMissingTag bool
}

func (w *Writer) writeConfig(out io.Writer, in interface{}) (int, error) {
	var n int

	return n, walker.Walk(
		in,
		func(f *walker.Field) error {
			s := w.Factory.Build(f.Field.Type)

			if s == nil {
				return nil
			}

			fks := walker.BuildFieldKeys(
				provider.WrapFullyQualifiedProvider(
					provider.NewStaticProvider("", nil, nil),
				),
				f,
				w.IgnoreMissingTag,
			)

			if len(fks) == 0 {
				return nil
			}

			if setter.IsUnmarshaler(f.Value.Type()) {
				return walker.SkipStruct
			}

			var b bytes.Buffer

			b.WriteString("\t- ")
			b.WriteString(fks[0])
			b.WriteString(": ")
			b.WriteString(s.String())

			if h := fieldHelp(f); h != "" {
				b.WriteString(" ")
				b.WriteString(h)
			}

			defaultValue := fieldDefault(f)
			providedKeys, tagDefault := w.providerKeys(f)

			if tagDefault != "" {
				defaultValue = tagDefault
			}

			if len(providedKeys) == 0 {
				return nil
			}

			if defaultValue != "" {
				b.WriteString(" (default: ")
				b.WriteString(defaultValue)
				b.WriteString(")")
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
}

func fieldDefault(f *walker.Field) string {
	fv := reflectutil.IndirectedValue(f.Value).FieldByName(f.Field.Name)

	if reflectutil.IsZero(fv) {
		return ""
	}

	v := reflectutil.IndirectedValue(fv).Interface()

	if ss, ok := v.(fmt.Stringer); ok {
		return ss.String()
	}

	return fmt.Sprintf("%+v", v)
}

func (w *Writer) providerKeys(f *walker.Field) ([]string, string) {
	var (
		providedKeys []string
		tagDefault   string
	)

	for _, p := range w.Providers {
		if _, ok := p.(dflt.Provider); ok {
			fqp := provider.WrapFullyQualifiedProvider(p)

			if ks := walker.BuildFieldKeys(fqp, f, w.IgnoreMissingTag); len(ks) > 0 {
				tagDefault = ks[0]
			}

			continue
		}

		fqp := provider.WrapFullyQualifiedProvider(p)

		ks := walker.BuildFieldKeys(fqp, f, w.IgnoreMissingTag)

		if len(ks) == 0 {
			continue
		}

		if kf, ok := p.(provider.KeyFormatter); ok {
			for i, k := range ks {
				ks[i] = kf.FormatKey(k)
			}
		}

		providedKeys = append(
			providedKeys,
			fmt.Sprintf("%s: %s", p.StructTag(), strings.Join(ks, ", ")),
		)
	}

	return providedKeys, tagDefault
}

func (w *Writer) Write(out io.Writer, ins ...interface{}) (int, error) {
	n, err := out.Write(defaultHeaders)

	if err != nil {
		return n, err
	}

	for _, in := range ins {
		nn, err := w.writeConfig(out, in)
		n += nn

		if err != nil {
			return n, err
		}
	}

	return n, nil
}
