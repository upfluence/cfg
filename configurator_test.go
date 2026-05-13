package cfg

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	dflt "github.com/upfluence/cfg/provider/default"
)

var (
	i64One   int64 = 1
	i64Two   int64 = 2
	i64Three int64 = 3

	errTest = errors.New("test test")
)

type mockProvider struct {
	st  map[string]string
	err error
}

func (p *mockProvider) StructTag() string { return "mock" }
func (p *mockProvider) Provide(_ context.Context, k string) (string, bool, error) {
	if p.err != nil {
		return "", false, p.err
	}

	v, ok := p.st[k]

	return v, ok, nil
}

type testCase struct {
	caseName string
	input    interface{}
	provider provider.Provider
	options  []Option

	dataAssertion func(*testing.T, interface{})
	errAssertion  func(*testing.T, error)
}

var noError = hasStaticError(nil)

func hasStaticError(out error) func(*testing.T, error) {
	return func(t *testing.T, in error) {
		if out != in {
			t.Errorf("Error returned: %v [ expected %v ]", in, out)
		}
	}
}

func hasError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Error exected ut none returned")
	}
}

type subStruct struct {
	string
}

func (ss *subStruct) Parse(v string) error {
	ss.string = v
	return nil
}

type errSubStruct struct{}

func (errSubStruct) Parse(string) error { return errTest }

type errValueStruct struct {
	A *errSubStruct
}

type sliceValueStruct struct {
	A []*subStruct
}

type valueStruct struct {
	A *subStruct
}

type valueTStruct struct {
	A subStruct
}

type mapValueStruct struct {
	A map[string]subStruct
}

type durationStruct struct {
	D time.Duration `mock:"d"`
}

type timeStruct struct {
	T time.Time `mock:"t"`
}

type floatStruct struct {
	F float64 `mock:"f"`
}

type float32Struct struct {
	F float32 `mock:"f"`
}

type uint64Struct struct {
	V uint64 `mock:"v"`
}

type mutiValuesStruct struct {
	Foo string `mock:"foo,bar,buz"`
}

func stringPtr(s string) *string { return &s }

type embedStruct1 struct {
	BasicStruct1
}

type BasicStruct1 struct {
	Fiz string
}

type basicStruct1 struct {
	Fiz string
}

type basicStruct2 struct {
	Foo int32
}

type basicStruct3 struct {
	Fiz *string
}

type basicStructBool struct {
	Bool bool `mock:"fzz"`
}

type nestedStruct struct {
	Nested struct {
		Inner *int64 `mock:"inner"`
	} `mock:"nested"`
}

type nestedV struct {
	Inner *int `mock:"inner"`
}

type nestedPtrStruct struct {
	Nested *nestedV `mock:"nested"`
}

type sliceStruct struct {
	Slice []int64 `mock:"slice"`
}

type slicePtrInt64Struct struct {
	Slice []*int64 `mock:"slice"`
}

type stringSliceStruct struct {
	Strings []string `mock:"strings"`
}

type mapStringIntStruct struct {
	Map map[string]int `mock:"map"`
}

type mapStringStringStruct struct {
	Map map[string]string `mock:"map"`
}

type mapStringStringsStruct struct {
	Map map[string][]string `mock:"map"`
}

type skipStruct struct {
	Foo string `mock:"-"`
}

type customString string

type customStringStruct struct {
	Foo customString `mock:"foo"`
}

type customInt int64

type customIntStruct struct {
	Foo customInt `mock:"foo"`
}

type customBool bool

type customBoolStruct struct {
	Foo customBool `mock:"foo"`
}

type marshalerStruct struct {
	Raw json.RawMessage
	IP  net.IP
}

func boolTestCase(in string, out bool) testCase {
	return testCase{
		input:         &basicStructBool{},
		caseName:      "basic-bool-value-" + in,
		provider:      &mockProvider{st: map[string]string{"fzz": in}},
		dataAssertion: deepEqual(&basicStructBool{out}),
		errAssertion:  noError,
	}
}

func deepEqual(x interface{}) func(*testing.T, interface{}) {
	return func(t *testing.T, y interface{}) {
		t.Helper()

		assert.Equalf(t, x, y, "Expected equality with %v but %v", x, y)
	}
}

func TestConfigurator(t *testing.T) {
	var (
		foo = "foo"
	)

	for _, tCase := range []testCase{
		testCase{
			caseName:      "basic-error",
			input:         &basicStruct1{},
			provider:      &mockProvider{err: errTest},
			dataAssertion: deepEqual(&basicStruct1{}),
			errAssertion:  hasError,
		},
		testCase{
			caseName:      "basic-no-ptr",
			input:         &basicStruct1{},
			provider:      &mockProvider{},
			dataAssertion: deepEqual(&basicStruct1{}),
			errAssertion:  noError,
		},
		testCase{
			input:         &embedStruct1{},
			caseName:      "basic-no-ptr-filled",
			provider:      &mockProvider{st: map[string]string{"Fiz": "Bar"}},
			dataAssertion: deepEqual(&embedStruct1{BasicStruct1: BasicStruct1{"Bar"}}),
			errAssertion:  noError,
		},
		testCase{
			input:         &embedStruct1{},
			caseName:      "basic-no-ptr-filled-ignore-missing--tag",
			options:       []Option{IgnoreMissingTag},
			provider:      &mockProvider{st: map[string]string{"Fiz": "Bar"}},
			dataAssertion: deepEqual(&embedStruct1{}),
			errAssertion:  noError,
		},
		testCase{
			input:         &basicStruct1{},
			caseName:      "basic-no-ptr-filled",
			provider:      &mockProvider{st: map[string]string{"Fiz": "Bar"}},
			dataAssertion: deepEqual(&basicStruct1{"Bar"}),
			errAssertion:  noError,
		},
		testCase{
			input:         &basicStruct2{},
			caseName:      "basic-int",
			provider:      &mockProvider{st: map[string]string{"Foo": "42"}},
			dataAssertion: deepEqual(&basicStruct2{42}),
			errAssertion:  noError,
		},
		testCase{
			input:         &customStringStruct{},
			caseName:      "typedef-string",
			provider:      &mockProvider{st: map[string]string{"foo": "bar"}},
			dataAssertion: deepEqual(&customStringStruct{Foo: "bar"}),
			errAssertion:  noError,
		},
		testCase{
			input:         &customIntStruct{},
			caseName:      "typedef-int",
			provider:      &mockProvider{st: map[string]string{"foo": "42"}},
			dataAssertion: deepEqual(&customIntStruct{Foo: 42}),
			errAssertion:  noError,
		},
		testCase{
			input:         &customBoolStruct{},
			caseName:      "typedef-bool",
			provider:      &mockProvider{st: map[string]string{"foo": "true"}},
			dataAssertion: deepEqual(&customBoolStruct{Foo: true}),
			errAssertion:  noError,
		},
		testCase{
			input:         &valueStruct{},
			caseName:      "basic-value",
			provider:      &mockProvider{st: map[string]string{"A": "foo"}},
			dataAssertion: deepEqual(&valueStruct{A: &subStruct{foo}}),
			errAssertion:  noError,
		},
		testCase{
			input:         &valueTStruct{},
			caseName:      "basic-t-value",
			provider:      &mockProvider{st: map[string]string{"A": "foo"}},
			dataAssertion: deepEqual(&valueTStruct{A: subStruct{"foo"}}),
			errAssertion:  noError,
		},
		testCase{
			input:         &errValueStruct{},
			caseName:      "basic-value-wrong",
			provider:      &mockProvider{st: map[string]string{"A": "foo"}},
			dataAssertion: deepEqual(&errValueStruct{}),
			errAssertion:  hasError,
		},
		testCase{
			input:    &sliceValueStruct{},
			caseName: "slice-value",
			provider: &mockProvider{st: map[string]string{"A": "foo,bar,buz"}},
			dataAssertion: deepEqual(
				&sliceValueStruct{A: []*subStruct{{"foo"}, {"bar"}, {"buz"}}},
			),
			errAssertion: noError,
		},
		testCase{
			input:    &mapValueStruct{},
			caseName: "map-value",
			provider: &mockProvider{st: map[string]string{"A": "foo=foo,bar=bar"}},
			dataAssertion: deepEqual(
				&mapValueStruct{A: map[string]subStruct{"foo": {"foo"}, "bar": {"bar"}}},
			),
			errAssertion: noError,
		},
		testCase{
			input:         &basicStruct2{},
			caseName:      "basic-int-wrong",
			dataAssertion: deepEqual(&basicStruct2{}),
			provider:      &mockProvider{st: map[string]string{"Foo": "dwadaw"}},
			errAssertion:  hasError,
		},
		testCase{
			caseName:      "basic-slice-int64",
			input:         &sliceStruct{},
			provider:      &mockProvider{st: map[string]string{"slice": "1,2,3"}},
			dataAssertion: deepEqual(&sliceStruct{Slice: []int64{1, 2, 3}}),
			errAssertion:  noError,
		},
		testCase{
			caseName: "basic-slice-ptr-int64",
			input:    &slicePtrInt64Struct{},
			provider: &mockProvider{st: map[string]string{"slice": "1,2,3"}},
			dataAssertion: deepEqual(&slicePtrInt64Struct{
				Slice: []*int64{&i64One, &i64Two, &i64Three},
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "basic-slice-string-slice",
			input:    &stringSliceStruct{},
			provider: &mockProvider{st: map[string]string{"strings": "foo,bar,buz"}},
			dataAssertion: deepEqual(&stringSliceStruct{
				Strings: []string{"foo", "bar", "buz"},
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "basic-map",
			input:    &mapStringIntStruct{},
			provider: &mockProvider{
				st: map[string]string{"map": "foo=1,bar=2,buz=3,fiz"},
			},
			dataAssertion: deepEqual(&mapStringIntStruct{
				Map: map[string]int{"foo": 1, "bar": 2, "buz": 3},
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "advanced-map",
			input:    &mapStringStringStruct{},
			provider: &mockProvider{
				st: map[string]string{"map": "foo='nested=k,v=z'"},
			},
			dataAssertion: deepEqual(&mapStringStringStruct{
				Map: map[string]string{"foo": "nested=k,v=z"},
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "advanced-map-2",
			input:    &mapStringStringsStruct{},
			provider: &mockProvider{
				st: map[string]string{"map": "foo='bar,buz',buz=biz"},
			},
			dataAssertion: deepEqual(&mapStringStringsStruct{
				Map: map[string][]string{"foo": {"bar", "buz"}, "buz": {"biz"}},
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "map-value-with-equals",
			input:    &mapStringStringStruct{},
			provider: &mockProvider{
				st: map[string]string{"map": "key=a=b=c"},
			},
			dataAssertion: deepEqual(&mapStringStringStruct{
				Map: map[string]string{"key": "a=b=c"},
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "duration",
			input:    &durationStruct{},
			provider: &mockProvider{st: map[string]string{"d": "5m"}},
			dataAssertion: deepEqual(&durationStruct{
				D: 5 * time.Minute,
			}),
			errAssertion: noError,
		},
		testCase{
			caseName: "time.Time",
			input:    &timeStruct{},
			provider: &mockProvider{
				st: map[string]string{"t": "2019-01-01T01:00:00"},
			},
			dataAssertion: deepEqual(
				&timeStruct{T: time.Date(2019, 1, 1, 1, 0, 0, 0, time.UTC)},
			),
			errAssertion: noError,
		},
		testCase{
			caseName:      "basic-ptr",
			input:         &basicStruct3{},
			provider:      &mockProvider{},
			dataAssertion: deepEqual(&basicStruct3{}),
			errAssertion:  noError,
		},
		testCase{
			input:         &basicStruct3{},
			caseName:      "basic-ptr-filled",
			provider:      &mockProvider{st: map[string]string{"Fiz": "Bar"}},
			dataAssertion: deepEqual(&basicStruct3{stringPtr("Bar")}),
			errAssertion:  noError,
		},
		testCase{
			input:         &basicStruct3{},
			caseName:      "basic-ptr-filled-empty",
			provider:      &mockProvider{st: map[string]string{"Fiz": ""}},
			dataAssertion: deepEqual(&basicStruct3{stringPtr("")}),
			errAssertion:  noError,
		},
		testCase{
			input:         &floatStruct{},
			caseName:      "basic-float",
			provider:      &mockProvider{st: map[string]string{"f": "0.5"}},
			dataAssertion: deepEqual(&floatStruct{F: .5}),
			errAssertion:  noError,
		},
		testCase{
			input:         &float32Struct{},
			caseName:      "basic-float32",
			provider:      &mockProvider{st: map[string]string{"f": "3.14"}},
			dataAssertion: deepEqual(&float32Struct{F: 3.14}),
			errAssertion:  noError,
		},
		testCase{
			input:         &uint64Struct{},
			caseName:      "large-uint64",
			provider:      &mockProvider{st: map[string]string{"v": "18446744073709551615"}},
			dataAssertion: deepEqual(&uint64Struct{V: 18446744073709551615}),
			errAssertion:  noError,
		},
		boolTestCase("t", true),
		boolTestCase("true", true),
		boolTestCase("1", true),
		boolTestCase("0", false),
		boolTestCase("f", false),
		boolTestCase("false", false),
		testCase{
			input:         &basicStructBool{},
			caseName:      "basic-bool-value-wrong",
			provider:      &mockProvider{st: map[string]string{"fzz": "blabla"}},
			dataAssertion: deepEqual(&basicStructBool{}),
			errAssertion:  hasError,
		},
		testCase{
			input:    &nestedStruct{},
			caseName: "nested_struct",
			provider: &mockProvider{st: map[string]string{"nested.inner": "123"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				if v := y.(*nestedStruct).Nested.Inner; *v != 123 {
					t.Errorf("Wrong result set: %v [ instead of: %v]", v, "fizz")
				}
			},
			errAssertion: noError,
		},
		testCase{
			input:    &nestedPtrStruct{},
			caseName: "nested_ptr_struct",
			provider: &mockProvider{st: map[string]string{"nested.inner": "123"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				if v := y.(*nestedPtrStruct).Nested.Inner; *v != 123 {
					t.Errorf("Wrong result set: %v [ instead of: %v]", v, "fizz")
				}
			},
			errAssertion: noError,
		},
		testCase{
			input:         &nestedPtrStruct{},
			caseName:      "nested_ptr_struct_no_values",
			provider:      &mockProvider{st: map[string]string{}},
			dataAssertion: deepEqual(&nestedPtrStruct{Nested: nil}),
			errAssertion:  noError,
		},

		testCase{
			input:    &mutiValuesStruct{},
			caseName: "multi value in tag",
			provider: &mockProvider{st: map[string]string{"fiz": "123"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				if v := y.(*mutiValuesStruct).Foo; v != "" {
					t.Errorf("Wrong result set: %v [ instead of: %v]", v, "")
				}
			},
			errAssertion: noError,
		},

		testCase{
			input:    &mutiValuesStruct{},
			caseName: "multi value in tag",
			provider: &mockProvider{st: map[string]string{"bar": "123"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				if v := y.(*mutiValuesStruct).Foo; v != "123" {
					t.Errorf("Wrong result set: %v [ instead of: %v]", v, "123")
				}
			},
			errAssertion: noError,
		},

		testCase{
			input:    &mutiValuesStruct{},
			caseName: "multi value in tag",
			provider: &mockProvider{st: map[string]string{"foo": "123"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				if v := y.(*mutiValuesStruct).Foo; v != "123" {
					t.Errorf("Wrong result set: %v [ instead of: %v]", v, "123")
				}
			},
			errAssertion: noError,
		},

		testCase{
			input:    &skipStruct{},
			caseName: "use - to skip providing",
			provider: &mockProvider{st: map[string]string{"Foo": "123"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				if v := y.(*skipStruct).Foo; v != "" {
					t.Errorf("Wrong result set: %v [ instead of: \"\"]", v)
				}
			},
			errAssertion: noError,
		},

		testCase{
			input:    &marshalerStruct{},
			caseName: "use the unmarshaller implementation",
			provider: &mockProvider{st: map[string]string{"Raw": "123.45", "IP": "127.1.2.3"}},
			dataAssertion: func(t *testing.T, y interface{}) {
				t.Log(y.(*marshalerStruct))
				if v := string(y.(*marshalerStruct).Raw); v != `"123.45"` {
					t.Errorf("Wrong result set: %v [ instead of: \"123.45\"]", v)
				}

				if v := y.(*marshalerStruct).IP.String(); v != "127.1.2.3" {
					t.Errorf("Wrong result set: %v [ instead of: \"127.1.2.3\"]", v)
				}
			},
			errAssertion: noError,
		},

		// Errors
		testCase{
			input:         nestedPtrStruct{},
			caseName:      "not_ptr_err",
			provider:      &mockProvider{},
			errAssertion:  hasStaticError(walker.ErrShouldBeAStructPtr),
			dataAssertion: func(t *testing.T, y interface{}) {},
		},
		testCase{
			input:         stringPtr("yolo"),
			provider:      &mockProvider{},
			caseName:      "not_struct_err",
			errAssertion:  hasStaticError(walker.ErrShouldBeAStructPtr),
			dataAssertion: func(t *testing.T, y interface{}) {},
		},
	} {
		t.Run(
			tCase.caseName,
			func(t *testing.T) {
				c := NewConfiguratorWithOptions(
					append(tCase.options, WithProviders(tCase.provider))...,
				)
				v := tCase.input

				tCase.errAssertion(t, c.Populate(context.Background(), v))
				tCase.dataAssertion(t, v)
			},
		)
	}
}

func TestDefaultProvider(t *testing.T) {
	for _, tc := range []struct {
		name string
		have any
		want any
	}{
		{
			name: "string default",
			have: &struct {
				Foo string `default:"bar"`
			}{},
			want: &struct {
				Foo string `default:"bar"`
			}{Foo: "bar"},
		},
		{
			name: "int default",
			have: &struct {
				V int `default:"42"`
			}{},
			want: &struct {
				V int `default:"42"`
			}{V: 42},
		},
		{
			name: "env overrides default",
			have: &struct {
				Foo string `default:"fallback" env:"TEST_DEFAULT_OVERRIDE"`
			}{},
			want: &struct {
				Foo string `default:"fallback" env:"TEST_DEFAULT_OVERRIDE"`
			}{Foo: "from_env"},
		},
		{
			name: "no default tag leaves zero value",
			have: &struct {
				Foo string
			}{},
			want: &struct {
				Foo string
			}{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "env overrides default" {
				t.Setenv("TEST_DEFAULT_OVERRIDE", "from_env")
			}

			err := NewDefaultConfigurator().Populate(context.Background(), tc.have)

			require.NoError(t, err)
			assert.Equal(t, tc.want, tc.have)
		})
	}
}

func TestHonorRequired(t *testing.T) {
	for _, tc := range []struct {
		name    string
		have    interface{}
		opts    []Option
		wantErr bool
	}{
		{
			name: "required field provided",
			have: &struct {
				Foo string `env:"FOO" required:"true"`
			}{},
			opts:    []Option{WithProviders(provider.NewStaticProvider("env", map[string]string{"FOO": "bar"}, nil))},
			wantErr: false,
		},
		{
			name: "required field missing",
			have: &struct {
				Foo string `env:"FOO" required:"true"`
			}{},
			opts:    []Option{},
			wantErr: true,
		},
		{
			name: "non-required field missing",
			have: &struct {
				Foo string `env:"FOO"`
			}{},
			opts:    []Option{},
			wantErr: false,
		},
		{
			name: "required false field missing",
			have: &struct {
				Foo string `env:"FOO" required:"false"`
			}{},
			opts:    []Option{},
			wantErr: false,
		},
		{
			name: "required field satisfied by default tag",
			have: &struct {
				Foo string `default:"bar" required:"true"`
			}{},
			opts:    []Option{WithProviders(dflt.Provider{})},
			wantErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewConfiguratorWithOptions(
				append(tc.opts, HonorRequired)...,
			)

			err := c.Populate(context.Background(), tc.have)

			if tc.wantErr {
				var re *RequiredError

				require.ErrorAs(t, err, &re)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type prefixedConfig struct {
	prefix []string
	value  any
}

func (p *prefixedConfig) WalkPrefix() []string { return p.prefix }
func (p *prefixedConfig) WalkValue() any       { return p.value }

type outerPrefixedConfig struct {
	Direct string `mock:"direct"`
	Nested *prefixedConfig
}

func TestPrefixedPopulate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		have     any
		provider provider.Provider
		want     any
	}{
		{
			name: "single segment prefix",
			have: &prefixedConfig{
				prefix: []string{"ns"},
				value:  &basicStruct1{},
			},
			provider: &mockProvider{st: map[string]string{"ns.Fiz": "bar"}},
			want:     &basicStruct1{Fiz: "bar"},
		},
		{
			name: "multi segment prefix",
			have: &prefixedConfig{
				prefix: []string{"foo", "bar"},
				value:  &basicStruct1{},
			},
			provider: &mockProvider{st: map[string]string{"foo.bar.Fiz": "baz"}},
			want:     &basicStruct1{Fiz: "baz"},
		},
		{
			name: "prefix with tagged field",
			have: &prefixedConfig{
				prefix: []string{"ns"},
				value:  &basicStructBool{},
			},
			provider: &mockProvider{st: map[string]string{"ns.fzz": "true"}},
			want:     &basicStructBool{Bool: true},
		},
		{
			name: "prefix with nested struct",
			have: &prefixedConfig{
				prefix: []string{"pfx"},
				value:  &nestedStruct{},
			},
			provider: &mockProvider{st: map[string]string{"pfx.nested.inner": "42"}},
			want: func() *nestedStruct {
				var v int64 = 42
				ns := &nestedStruct{}

				ns.Nested.Inner = &v

				return ns
			}(),
		},
		{
			name: "empty prefix behaves like normal populate",
			have: &prefixedConfig{
				prefix: nil,
				value:  &basicStruct1{},
			},
			provider: &mockProvider{st: map[string]string{"Fiz": "val"}},
			want:     &basicStruct1{Fiz: "val"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewConfiguratorWithOptions(WithProviders(tc.provider))

			err := c.Populate(context.Background(), tc.have)

			require.NoError(t, err)
			assert.Equal(t, tc.want, tc.have.(*prefixedConfig).value)
		})
	}
}

func TestNestedPrefixedPopulate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		have      *outerPrefixedConfig
		provider  provider.Provider
		wantOuter string
		wantInner *basicStruct1
	}{
		{
			name: "nested prefixed field",
			have: &outerPrefixedConfig{
				Nested: &prefixedConfig{
					prefix: []string{"ns"},
					value:  &basicStruct1{},
				},
			},
			provider:  &mockProvider{st: map[string]string{"direct": "top", "Nested.ns.Fiz": "deep"}},
			wantOuter: "top",
			wantInner: &basicStruct1{Fiz: "deep"},
		},
		{
			name: "nested prefixed with multi-segment prefix",
			have: &outerPrefixedConfig{
				Nested: &prefixedConfig{
					prefix: []string{"a", "b"},
					value:  &basicStruct1{},
				},
			},
			provider:  &mockProvider{st: map[string]string{"Nested.a.b.Fiz": "val"}},
			wantOuter: "",
			wantInner: &basicStruct1{Fiz: "val"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewConfiguratorWithOptions(WithProviders(tc.provider))

			err := c.Populate(context.Background(), tc.have)

			require.NoError(t, err)
			assert.Equal(t, tc.wantOuter, tc.have.Direct)
			assert.Equal(t, tc.wantInner, tc.have.Nested.value)
		})
	}
}

func ExampleNewDefaultConfigurator() {
	os.Setenv("FOO", "bar")
	cfg := struct {
		Foo string `env:"FOO"`
	}{}

	if err := NewDefaultConfigurator().Populate(
		context.Background(),
		&cfg,
	); err != nil {
		os.Exit(1)
	}

	fmt.Println(cfg.Foo)
	// Output:
	// bar
}
