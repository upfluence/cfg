package setters

import (
	"fmt"
	"math"
	"reflect"
)

type (
	intTransformer   func(int64, bool) (interface{}, error)
	floatTransformer func(float64, bool) (interface{}, error)
)

func reflectIntTransformer(v int64, ptr bool,
	fun func(int64) (interface{}, error)) (interface{}, error) {

	if ret, err := fun(v); err != nil {
		return nil, err
	} else {
		if ptr {
			return &ret, nil
		}
		return ret, nil
	}
}

func reflectFloatTransformer(v float64, ptr bool,
	fun func(float64) (interface{}, error)) (interface{}, error) {

	if ret, err := fun(v); err != nil {
		return nil, err
	} else {
		if ptr {
			return &ret, nil
		}
		return ret, nil
	}
}

var (
	intTransformers = map[reflect.Kind]intTransformer{
		reflect.Int: func(v int64, ptr bool) (interface{}, error) {
			return reflectIntTransformer(v, ptr, intFuncs["int"])
		},
		reflect.Int64: func(v int64, ptr bool) (interface{}, error) {
			return reflectIntTransformer(v, ptr, intFuncs["int64"])
		},
		reflect.Int32: func(v int64, ptr bool) (interface{}, error) {
			return reflectIntTransformer(v, ptr, intFuncs["int32"])
		},
		reflect.Int16: func(v int64, ptr bool) (interface{}, error) {
			return reflectIntTransformer(v, ptr, intFuncs["int16"])
		},
		reflect.Int8: func(v int64, ptr bool) (interface{}, error) {
			return reflectIntTransformer(v, ptr, intFuncs["int8"])
		},
	}

	floatTransformers = map[reflect.Kind]floatTransformer{
		reflect.Float64: func(v float64, ptr bool) (interface{}, error) {
			return reflectFloatTransformer(v, ptr, floatFuncs["float64"])
		},
		reflect.Float32: func(v float64, ptr bool) (interface{}, error) {
			return reflectFloatTransformer(v, ptr, floatFuncs["float32"])
		},
	}
)

var (
	intFuncs = map[string]func(int64) (interface{}, error){
		"int": func(v int64) (interface{}, error) {
			const (
				MaxUint  = ^uint(0)
				MaxRange = int64(int(MaxUint >> 1))
				MinRange = int64(-int(MaxRange) - 1)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int8)", v)
			}

			return int(v), nil
		},
		"int8": func(v int64) (interface{}, error) {
			const (
				MinRange = int64(math.MinInt8)
				MaxRange = int64(math.MaxInt8)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int8)", v)
			}

			return int8(v), nil
		},
		"int16": func(v int64) (interface{}, error) {
			const (
				MinRange = int64(math.MinInt16)
				MaxRange = int64(math.MaxInt16)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int16)", v)
			}

			return int16(v), nil
		},
		"int32": func(v int64) (interface{}, error) {
			const (
				MinRange = int64(math.MinInt32)
				MaxRange = int64(math.MaxInt32)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int32)", v)
			}

			return int32(v), nil
		},
		"int64": func(v int64) (interface{}, error) {
			return v, nil
		},
	}

	floatFuncs = map[string]func(float64) (interface{}, error){
		"float32": func(v float64) (interface{}, error) {
			if float64(math.MaxFloat32) < math.Abs(v) {
				return v, fmt.Errorf(
					"floatTransformers error: range -> %f (reflect.Float32)", v)
			}

			return float32(v), nil
		},
		"float64": func(v float64) (interface{}, error) {
			return v, nil
		},
	}
)
