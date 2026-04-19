package walker

import (
	"reflect"
	"strings"

	"github.com/upfluence/cfg/provider"
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

func buildStructFieldKey(p provider.FullyQualifiedProvider, sf reflect.StructField, ignoreMissingTag bool) ([]string, bool) {
	if t := p.StructTag(); t != "" {
		switch v, ok := sf.Tag.Lookup(t); v {
		case "":
			if ok {
				return []string{}, true
			}

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

	dfv := p.DefaultFieldValue(sf.Name)

	if dfv == "" {
		return nil, true
	}

	return []string{dfv}, true
}

func BuildFieldKeys(p provider.FullyQualifiedProvider, f *Field, ignoreMissingTag bool) []string {
	var fss [][]string

	if ok := walkFields(f, func(sf reflect.StructField) bool {
		fs, ok := buildStructFieldKey(p, sf, ignoreMissingTag)

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
		if p.DefaultFieldValue("") == "" {
			return nil
		}

		return []string{"config"}
	}

	return joinPermutation(fss, p.JoinFieldKeys)
}

func joinPermutation(fss [][]string, join func(string, string) string) []string {
	switch len(fss) {
	case 0:
		return nil
	case 1:
		return fss[0]
	}

	left := fss[0]
	right := joinPermutation(fss[1:], join)

	var res []string

	for _, l := range left {
		for _, r := range right {
			res = append(res, join(l, r))
		}
	}

	return res
}
