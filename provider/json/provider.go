package json

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/upfluence/cfg/provider"
)

var ErrJSONMalformated = errors.New("cfg/provider/json: Payload not formatted correctly")

type Provider struct {
	store map[string]interface{}
}

func NewProviderFromReader(r io.Reader) provider.Provider {
	var v = make(map[string]interface{})

	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return provider.ProvideError("json", err)
	}

	return &Provider{store: v}
}

func (*Provider) StructTag() string { return "json" }

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	var (
		cur         = p.store
		splittedKey = strings.Split(v, ".")

		res interface{}
	)

	for i, k := range splittedKey {
		t := cur[k]

		if t == nil {
			return "", false, nil
		}

		if i == len(splittedKey)-1 {
			res = t
			continue
		}

		res, ok := t.(map[string]interface{})

		if !ok {
			return "", false, ErrJSONMalformated
		}

		cur = res
	}

	return fmt.Sprintf("%v", res), true, nil
}
