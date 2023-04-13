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
	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
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
	Foo int64
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
