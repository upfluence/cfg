package cfg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/upfluence/cfg/provider"
)

var (
	i64One   int64 = 1
	i64Two   int64 = 2
	i64Three int64 = 3
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
		t.Errorf("Error returned but none returned")
	}
}

func stringPtr(s string) *string { return &s }

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
		if !reflect.DeepEqual(x, y) {
			t.Errorf("Expected equality with %v but %v", x, y)
		}
	}
}

func TestConfigurator(t *testing.T) {
	var errTest = errors.New("test test")

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
			provider: &mockProvider{st: map[string]string{"map": "foo=1,bar=2,buz=3,fiz"}},
			dataAssertion: deepEqual(&mapStringIntStruct{
				Map: map[string]int{"foo": 1, "bar": 2, "buz": 3},
			}),
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

		// Errors
		testCase{
			input:         nestedPtrStruct{},
			caseName:      "not_ptr_err",
			errAssertion:  hasStaticError(ErrShouldBeAStructPtr),
			dataAssertion: func(t *testing.T, y interface{}) {},
		},
		testCase{
			input:         stringPtr("yolo"),
			caseName:      "not_struct_err",
			errAssertion:  hasStaticError(ErrShouldBeAStructPtr),
			dataAssertion: func(t *testing.T, y interface{}) {},
		},
	} {
		t.Run(
			tCase.caseName,
			func(t *testing.T) {
				c := NewConfigurator(tCase.provider)
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
