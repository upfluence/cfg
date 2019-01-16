package cfg

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/upfluence/cfg/internal/setters"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

var ErrShouldBeAStructPtr = errors.New("cfg: input should be a pointer")

type Configurator interface {
	Populate(context.Context, interface{}) error
}

type configurator struct {
	providers []provider.Provider
	factory   setters.SetterFactory
}

func NewDefaultConfigurator(providers ...provider.Provider) Configurator {
	return NewConfigurator(
		append(providers, env.NewDefaultProvider(), flags.NewDefaultProvider())...,
	)
}

func NewConfigurator(providers ...provider.Provider) Configurator {
	return &configurator{providers: providers, factory: &defaultSetterFactory{}}
}

func (c *configurator) Populate(ctx context.Context, out interface{}) error {
	vVal := reflect.ValueOf(out)

	if vVal.Type().Kind() != reflect.Ptr {
		return ErrShouldBeAStructPtr
	}

	indirectVType := vVal.Type().Elem()

	if indirectVType.Kind() != reflect.Struct {
		return ErrShouldBeAStructPtr
	}

	for _, p := range c.providers {
		if err := c.populate(ctx, p, vVal, nil); err != nil {
			return err
		}
	}

	return nil
}

func (c *configurator) populate(ctx context.Context, p provider.Provider, vVal reflect.Value, ns []string) error {
	var indirectVType = indirectedType(vVal.Type())

	for i := 0; i < indirectVType.NumField(); i++ {
		field := indirectVType.Field(i)
		s := c.factory.buildSetter(field)
		v := indirectedValue(vVal).FieldByName(field.Name)
		n := field.Name

		if t := p.StructTag(); t != "" {
			if v, ok := field.Tag.Lookup(t); ok && v != "" {
				n = v
			}
		}

		if !v.CanSet() {
			continue
		}

		if s == nil && indirectedType(field.Type).Kind() == reflect.Struct {
			if field.Type.Kind() != reflect.Ptr {
				v = v.Addr()
			} else {
				v.Set(reflect.New(field.Type.Elem()))
			}

			if n != "" {
				ns = append(ns, n)
			}

			c.populate(ctx, p, v, ns)
		} else if s != nil {
			v, ok, err := p.Provide(ctx, strings.Join(append(ns, n), "."))

			if err != nil {
				return errors.Wrapf(
					err,
					"Populate {struct: %T field: %s}",
					v,
					field.Name,
				)
			}

			if !ok {
				continue
			}

			if err := s.set(v, vVal.Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}
