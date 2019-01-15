package setters

import (
	"fmt"
	"math"
	"reflect"

	"github.com/pkg/errors"
)

type (
	intTransformer   func(int64, bool) (interface{}, error)
	floatTransformer func(float64, bool) (interface{}, error)
)

func reflectIntTransformer(
	v int64, ptr bool, kind reflect.Kind) (interface{}, error) {

	fun := intFuncs(kind)
	if ret, err := fun(v); err != nil {
		return nil, err
	} else {
		if ptr {
			return &ret, nil
		}
		return ret, nil
	}
}

func reflectFloatTransformer(
	v float64, ptr bool, kind reflect.Kind) (interface{}, error) {

	fun := floatFuncs(kind)
	if ret, err := fun(v); err != nil {
		return nil, err
	} else {
		if ptr {
			return &ret, nil
		}
		return ret, nil
	}
}

func intTransformerFactory(t reflect.Kind) intTransformer {
	return func(v int64, ptr bool) (interface{}, error) {
		return reflectIntTransformer(v, ptr, t)
	}
}

func floatTransformerFactory(t reflect.Kind) floatTransformer {
	return func(v float64, ptr bool) (interface{}, error) {
		return reflectFloatTransformer(v, ptr, t)
	}
}

func intFuncs(kind reflect.Kind) func(int64) (interface{}, error) {
	switch kind {
	case reflect.Int:
		return func(v int64) (interface{}, error) {
			const (
				MaxUint  = ^uint(0)
				MaxRange = int64(int(MaxUint >> 1))
				MinRange = int64(-int(MaxRange) - 1)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int8)",
					v,
				)
			}

			return int(v), nil
		}
	case reflect.Int8:
		return func(v int64) (interface{}, error) {
			const (
				MinRange = int64(math.MinInt8)
				MaxRange = int64(math.MaxInt8)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int8)",
					v,
				)
			}

			return int8(v), nil
		}
	case reflect.Int16:
		return func(v int64) (interface{}, error) {
			const (
				MinRange = int64(math.MinInt16)
				MaxRange = int64(math.MaxInt16)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int16)",
					v,
				)
			}

			return int16(v), nil
		}
	case reflect.Int32:
		return func(v int64) (interface{}, error) {
			const (
				MinRange = int64(math.MinInt32)
				MaxRange = int64(math.MaxInt32)
			)
			if v > MaxRange || v < MinRange {
				return nil, fmt.Errorf(
					"intTransformer error: range -> %d (reflect.Int32)",
					v,
				)
			}

			return int32(v), nil
		}
	case reflect.Int64:
		return func(v int64) (interface{}, error) {
			return v, nil
		}
	default:
		return func(v int64) (interface{}, error) {
			return nil, errors.New(
				"INTERNAL ERROR: method transformer not implemented.",
			)
		}
	}

}

func floatFuncs(kind reflect.Kind) func(float64) (interface{}, error) {
	switch kind {
	case reflect.Float32:
		return func(v float64) (interface{}, error) {
			if float64(math.MaxFloat32) < math.Abs(v) {
				return v, fmt.Errorf(
					"floatTransformers error: range -> %f (reflect.Float32)",
					v,
				)
			}

			return float32(v), nil
		}
	case reflect.Float64:
		return func(v float64) (interface{}, error) {
			return v, nil
		}
	default:
		return func(v float64) (interface{}, error) {
			return nil, errors.New(
				"INTERNAL ERROR: method transformer not implemented.",
			)
		}
	}
}
