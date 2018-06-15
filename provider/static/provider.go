package static

import (
	"bytes"
	stdjson "encoding/json"

	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/json"
)

func NewProvider(d interface{}) (provider.Provider, error) {
	var buf, err = stdjson.Marshal(d)

	if err != nil {
		return nil, err
	}

	return json.NewProviderFromReader(bytes.NewReader(buf))
}
