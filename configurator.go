package cfg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

type Configurator interface {
	Populate(context.Context, interface{}) error
}

type configurator struct {
	providers []provider.Provider
	factory   setterFactory

	output   io.Writer
	helpKeys map[string][]string
}

func NewDefaultConfigurator(providers ...provider.Provider) Configurator {
	return NewConfigurator(
		append(providers, env.NewDefaultProvider(), flags.NewDefaultProvider())...,
	)
}

func NewConfigurator(providers ...provider.Provider) Configurator {
	return &configurator{
		providers: providers,
		factory:   &defaultSetterFactory{},
		output:    os.Stderr,
		helpKeys: map[string][]string{
			"flag": []string{"h", "help"},
			"env":  []string{"HELP"},
		},
	}
}

func (c *configurator) Populate(ctx context.Context, in interface{}) error {
	switch err := c.parseHelp(ctx); err {
	case ErrHelp:
		c.PrintDefaults(in)
		os.Exit(2)
	case nil:
	default:
		return err
	}

	return walker.Walk(
		in,
		func(f *walker.Field) error { return c.walkFunc(ctx, f) },
	)
}

func (c *configurator) parseHelp(ctx context.Context) error {
	var bp boolParser

	for _, p := range c.providers {
		ks := c.helpKeys[p.StructTag()]

		for _, k := range ks {
			v, ok, err := p.Provide(ctx, k)

			if err != nil {
				return err
			}

			if !ok {
				continue
			}

			help, err := bp.parse(v, false)

			if err != nil {
				return err
			}

			if v, ok := help.(bool); ok && v {
				return ErrHelp
			}
		}
	}

	return nil
}

var (
	ErrHelp = errors.New("help required")

	defaultHeaders = []byte("Arguments:\n")
)

func (c *configurator) PrintDefaults(in interface{}) error {
	c.output.Write(defaultHeaders)

	return walker.Walk(
		in,
		func(f *walker.Field) error {
			s := c.factory.buildSetter(f.Field)

			if s == nil {
				return nil
			}

			var b bytes.Buffer

			b.WriteString("\t- ")

			fn := buildFieldName(f)
			b.WriteString(fn)

			b.WriteString(": ")
			b.WriteString(s.String())

			if fv := indirectedValue(f.Value).FieldByName(f.Field.Name); !isZero(fv) {
				v := indirectedValue(fv).Interface()

				b.WriteString(" (default: ")

				if ss, ok := v.(fmt.Stringer); ok {
					b.WriteString(ss.String())
				} else {
					fmt.Fprintf(&b, "%+v", v)
				}

				b.WriteString(")")
			}

			b.WriteTo(c.output)

			return nil
		},
	)
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface:
		return v.IsNil()
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}

		return isZero(v.Elem())
	}

	return false
}

func (c *configurator) walkFunc(ctx context.Context, f *walker.Field) error {
	s := c.factory.buildSetter(f.Field)

	if s == nil {
		return nil
	}

	for _, p := range c.providers {
		v, ok, err := p.Provide(ctx, buildFieldKey(p, f))

		if err != nil {
			return errors.Wrapf(
				err,
				"Populate {struct: %T field: %s}",
				v,
				f.Field.Name,
			)
		}

		if !ok {
			continue
		}

		if err := s.set(v, f.Value.Interface()); err != nil {
			return err
		}
	}

	return nil
}

func walkFields(f *walker.Field, fn func(reflect.StructField)) {
	var (
		fs = []reflect.StructField{f.Field}
		a  = f.Ancestor
	)

	for a != nil {
		fs = append(fs, a.Field)
		a = a.Ancestor
	}

	for i := len(fs); i > 0; i-- {
		fn(fs[i-1])
	}
}

func buildStructFieldKey(p provider.Provider, sf reflect.StructField) string {
	if v, ok := sf.Tag.Lookup(p.StructTag()); ok {
		return v
	}

	return sf.Name

}

func buildFieldKey(p provider.Provider, f *walker.Field) string {
	var fs []string

	walkFields(f, func(sf reflect.StructField) {
		fs = append(fs, buildStructFieldKey(p, sf))
	})

	return strings.Join(fs, ".")
}

func buildFieldName(f *walker.Field) string {
	var fs []string

	walkFields(f, func(sf reflect.StructField) { fs = append(fs, sf.Name) })

	return strings.Join(fs, ".")
}
