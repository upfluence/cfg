package static

import (
	"bytes"
	"encoding/json"

	"github.com/upfluence/cfg/provider"
	pjson "github.com/upfluence/cfg/provider/json"
)

func NewProvider(d interface{}) (provider.Provider, error) {
	var (
		buf bytes.Buffer

		enc = json.NewEncoder(&buf)
	)

	if err := enc.Encode(d); err != nil {
		return nil, err
	}

	return pjson.NewProviderFromReader(&buf)
}
