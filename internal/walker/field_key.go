package walker

import (
	"reflect"
	"strings"
)

func walkFields(f *Field, fn func(reflect.StructField) bool) bool {
	var (
		fs = []reflect.StructField{f.Field}
		a  = f.Ancestor
	)

	for a != nil {
		fs = append(fs, a.Field)
		a = a.Ancestor
	}

	for i := len(fs); i > 0; i-- {
		if ok := fn(fs[i-1]); !ok {
			return false
		}
	}

	return true
}

func buildStructFieldKey(t string, sf reflect.StructField, ignoreMissingTag bool) ([]string, bool) {
	if t != "" {
		switch v, _ := sf.Tag.Lookup(t); v {
		case "":
			if ignoreMissingTag {
				return nil, false
			}
		case "-":
			return nil, false
		default:
			return strings.Split(v, ","), true
		}
	}

	if sf.Anonymous {
		return nil, true
	}

	return []string{sf.Name}, true
}

func BuildFieldKeys(t string, f *Field, ignoreMissingTag bool) []string {
	var fss [][]string

	if ok := walkFields(f, func(sf reflect.StructField) bool {
		fs, ok := buildStructFieldKey(t, sf, ignoreMissingTag)

		if !ok {
			return false
		}

		if len(fs) > 0 {
			fss = append(fss, fs)
		}

		return true
	}); !ok {
		return nil
	}

	if len(fss) == 0 {
		return []string{"config"}
	}

	return joinPermutation(fss, ".")
}

func joinPermutation(fss [][]string, delim string) []string {
	switch len(fss) {
	case 0:
		return nil
	case 1:
		return fss[0]
	}

	left := fss[0]
	right := joinPermutation(fss[1:], delim)

	var res []string

	for _, l := range left {
		for _, r := range right {
			res = append(res, strings.Join([]string{l, r}, delim))
		}
	}

	return res
}
