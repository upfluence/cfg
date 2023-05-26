package setter

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/stringutil"
)

var (
	durationType        = reflect.TypeOf(time.Duration(0))
	timeType            = reflect.TypeOf(time.Time{})
	valueType           = reflect.TypeOf((*Value)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

	durationParser = &staticParser{t: "duration", fn: parseDuration}
	boolParser     = &staticParser{t: "bool", fn: parseBool}
	stringParser   = &staticParser{t: "string", fn: parseString}

	presetParsers = map[reflect.Type]func(factoryOptions) parser{
		durationType: func(factoryOptions) parser { return durationParser },
		timeType: func(fo factoryOptions) parser {
			return timeParser{opts: fo.tpo}
		},
	}

	interfaceParsers = map[reflect.Type]func(interface{}, string) error{
		valueType:           assignValue,
		textUnmarshalerType: assignTextUnmarshaler,
		jsonUnmarshalerType: assignJSONUnmarshaler,
	}

	defaultFactoryOptions = factoryOptions{tpo: defaultTimeParserOption}

	DefaultFactory = NewDefaultFactory()
)

func IsUnmarshaler(t reflect.Type) bool {
	for it := range interfaceParsers {
		if t.Implements(it) {
			return true
		}
	}

	if t.Kind() != reflect.Ptr {
		return IsUnmarshaler(reflect.PtrTo(t))
	}

	return false
}

type factoryOptions struct {
	tpo timeParserOption
}

func WithDateFormat(fmt string) FactoryOption {
	return func(opts *factoryOptions) { opts.tpo.dateFmt = fmt }
}

type FactoryOption func(*factoryOptions)

type Value interface {
	Parse(string) error
}

type Factory interface {
	Build(reflect.Type) Setter
}

type defaultFactory struct {
	opts factoryOptions
}

func NewDefaultFactory(opts ...FactoryOption) Factory {
	var o = defaultFactoryOptions

	for _, opt := range opts {
		opt(&o)
	}

	return &defaultFactory{opts: o}
}

func (df *defaultFactory) buildBasicParser(t reflect.Type) (parser, bool) {
	var (
		k  = t.Kind()
		pt = t

		ptr bool
	)

	if k == reflect.Ptr {
		k = t.Elem().Kind()
		ptr = true
		t = t.Elem()
	} else {
		pt = reflect.PtrTo(t)
	}

	if p, ok := presetParsers[t]; ok {
		return p(df.opts), ptr
	}

	for it, fn := range interfaceParsers {
		if pt.Implements(it) {
			return &interfaceParser{t: t, fn: fn}, ptr
		}
	}

	switch {
	case k == reflect.String:
		return stringParser, ptr
	case k >= reflect.Int && k <= reflect.Uint64:
		return &intParser{transformer: intTransformers[k]}, ptr
	case k == reflect.Float32 || k == reflect.Float64:
		return floatParser(k), ptr
	case k == reflect.Bool:
		return boolParser, ptr
	}

	return nil, false
}

func (df *defaultFactory) buildParser(t reflect.Type) (parser, bool) {
	k := t.Kind()

	switch k {
	case reflect.Slice:
		// Make sure it is not a []byte
		if t.Elem().Kind() != reflect.Uint8 {
			p, ptr := df.buildBasicParser(t.Elem())

			if p == nil {
				return nil, false
			}

			return &sliceParser{p: p, t: t, ptr: ptr}, false
		}
	case reflect.Map:
		vp, vptr := df.buildParser(t.Elem())

		if vp == nil {
			return nil, false
		}

		kp, kptr := df.buildBasicParser(t.Key())

		if kp == nil {
			return nil, false
		}

		return &mapParser{t: t, vp: vp, vptr: vptr, kp: kp, kptr: kptr}, false
	}

	return df.buildBasicParser(t)
}

func (df *defaultFactory) Build(t reflect.Type) Setter {
	if p, _ := df.buildParser(reflectutil.IndirectedType(t)); p != nil {
		return &parserSetter{parser: p}
	}

	return nil
}

type Setter interface {
	fmt.Stringer

	Set(string, reflect.Value) error
}

type NotImplementedError struct {
	field reflect.StructField
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("cfg: Setter not implemented for type %v", e.field.Type)
}

type parserSetter struct {
	parser parser
}

func (s *parserSetter) String() string { return s.parser.String() }

func (s *parserSetter) Set(value string, t reflect.Value) error {
	v, err := s.parser.parse(value, t.Type().Kind() == reflect.Ptr)

	if err != nil {
		return err
	}

	t.Set(reflect.ValueOf(v))

	return nil
}

type parser interface {
	fmt.Stringer

	parse(string, bool) (interface{}, error)
}

type interfaceParser struct {
	t reflect.Type

	fn func(interface{}, string) error
}

func (ip *interfaceParser) String() string { return ip.t.String() }

func (ip *interfaceParser) parse(v string, ptr bool) (interface{}, error) {
	rv := reflect.New(ip.t)

	if err := ip.fn(rv.Interface(), v); err != nil {
		return nil, err
	}

	if ptr {
		return rv.Interface(), nil
	}

	return rv.Elem().Interface(), nil
}

func assignValue(v interface{}, txt string) error {
	return v.(Value).Parse(txt)
}

func assignTextUnmarshaler(v interface{}, txt string) error {
	return v.(encoding.TextUnmarshaler).UnmarshalText([]byte(txt))
}

func shouldQuote(txt string) bool {
	if len(txt) < 2 {
		return true
	}

	if txt[0] == '[' && txt[len(txt)-1] == ']' {
		return false
	}

	if txt[0] == '{' && txt[len(txt)-1] == '}' {
		return false
	}

	return true
}

func assignJSONUnmarshaler(v interface{}, txt string) error {
	if shouldQuote(txt) {
		unquoted, err := strconv.Unquote(txt)

		if err != nil {
			unquoted = txt
		}

		txt = strconv.Quote(unquoted)
	}

	return v.(json.Unmarshaler).UnmarshalJSON([]byte(txt))
}

type mapParser struct {
	t reflect.Type

	vp, kp parser

	vptr, kptr bool
}

func (mp *mapParser) String() string {
	return fmt.Sprintf("map[%s]%s", mp.kp.String(), mp.vp.String())
}

func (mp *mapParser) parse(v string, ptr bool) (interface{}, error) {
	args, err := stringutil.Split(v, ',')

	if err != nil {
		return nil, errors.Wrapf(err, "%q is not a correct map value", v)
	}

	res := reflect.MakeMap(mp.t)

	for _, arg := range args {
		vs, err := stringutil.Split(arg, '=')

		if err != nil {
			return nil, errors.Wrapf(err, "%q is not a correct key/value clause", v)
		}

		if len(vs) != 2 {
			continue
		}

		k, err := mp.kp.parse(vs[0], mp.kptr)

		if err != nil {
			return nil, err
		}

		v, err := mp.vp.parse(vs[1], mp.vptr)

		if err != nil {
			return nil, err
		}

		res.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}

	return res.Interface(), nil
}

type sliceParser struct {
	t reflect.Type

	p   parser
	ptr bool
}

func (sp *sliceParser) String() string {
	return fmt.Sprintf("[]%s", sp.p.String())
}

func (sp *sliceParser) parse(v string, ptr bool) (interface{}, error) {
	args, err := stringutil.Split(v, ',')

	if err != nil {
		return nil, errors.Wrapf(err, "%q is not a correct slice value", v)
	}

	res := reflect.MakeSlice(sp.t, 0, len(args))

	for _, arg := range args {
		v, err := sp.p.parse(arg, sp.ptr)

		if err != nil {
			return nil, err
		}

		res = reflect.Append(res, reflect.ValueOf(v))
	}

	return res.Interface(), nil
}

type intTransformer func(int64, bool) interface{}

type floatParser reflect.Kind

func (fp floatParser) String() string { return "float" }

func (fp floatParser) parse(value string, ptr bool) (interface{}, error) {
	var v, err = strconv.ParseFloat(value, 64)

	if err != nil {
		return nil, err
	}

	if ptr {
		if reflect.Kind(fp) == reflect.Float32 {
			vv := float32(v)
			return &vv, nil
		}

		return &v, nil
	}

	if reflect.Kind(fp) == reflect.Float32 {
		return float32(fp), nil
	}

	return v, nil
}

type intParser struct {
	transformer intTransformer
}

func (*intParser) String() string { return "integer" }

func (s *intParser) parse(value string, ptr bool) (interface{}, error) {
	var v, err = strconv.ParseInt(value, 10, 0)

	if err != nil {
		return nil, err
	}

	return s.transformer(v, ptr), nil
}

type timeParserOption struct {
	dateFmt string
}

var defaultTimeParserOption = timeParserOption{dateFmt: "2006-01-02T15:04:05"}

type timeParser struct {
	opts timeParserOption
}

func (timeParser) String() string { return "time" }

func (tp timeParser) parse(value string, ptr bool) (interface{}, error) {
	t, err := time.Parse(tp.opts.dateFmt, value)

	if err != nil {
		return nil, err
	}

	if ptr {
		return &t, nil
	}

	return t, nil
}

type staticParser struct {
	t string

	fn func(string, bool) (interface{}, error)
}

func (sp *staticParser) String() string { return sp.t }
func (sp *staticParser) parse(value string, ptr bool) (interface{}, error) {
	return sp.fn(value, ptr)
}

func parseDuration(value string, ptr bool) (interface{}, error) {
	d, err := time.ParseDuration(value)

	if err != nil {
		return nil, err
	}

	if ptr {
		return &d, nil
	}

	return d, nil
}

type NotBoolValueError struct {
	value string
}

func (e *NotBoolValueError) Error() string {
	return fmt.Sprintf("cfg: Can't parse %q in a bool value", e.value)
}

func parseBool(value string, ptr bool) (interface{}, error) {
	var v bool

	switch strings.TrimSpace(value) {
	case "t", "1", "true", "yes", "y":
		v = true
	case "f", "0", "false", "no", "n":
	default:
		return nil, &NotBoolValueError{value: value}
	}

	if ptr {
		return &v, nil
	}

	return v, nil
}

func parseString(v string, ptr bool) (interface{}, error) {
	if ptr {
		x := v
		return &x, nil
	}

	return v, nil
}
