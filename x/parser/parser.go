package parser

import (
	"errors"
	"reflect"

	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/setter"
)

var (
	ErrShouldBePtr = errors.New("x/parser: input should be a pointer")

	WithDateFormat = setter.WithDateFormat
)

type Option = setter.FactoryOption

type Parser struct {
	sf setter.Factory
}

func NewParser(opts ...Option) *Parser {
	return &Parser{sf: setter.NewDefaultFactory(opts...)}
}

func (p *Parser) Parse(data string, target interface{}) error {
	v := reflect.ValueOf(target)

	if v.Kind() != reflect.Ptr {
		return ErrShouldBePtr
	}

	return p.sf.Build(v.Type()).Set(data, reflectutil.IndirectedValue(v))
}

func Parse(data string, target interface{}, opts ...Option) error {
	return NewParser(opts...).Parse(data, target)
}
