package cfg

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type helpStructConfig struct {
	Yolo string `help:"this is the help message" flag:"yolo,y"`
}

func TestPrintDefaults(t *testing.T) {
	for _, tt := range []struct {
		in  interface{}
		out string
	}{
		{
			in:  &mapStringIntStruct{},
			out: "Arguments:\n\t- Map: map[string]integer (env: MAP, flag: --map)\n",
		},
		{
			in:  &mapStringIntStruct{Map: map[string]int{"fiz": 42}},
			out: "Arguments:\n\t- Map: map[string]integer (default: map[fiz:42]) (env: MAP, flag: --map)\n",
		},
		{
			in:  &nestedPtrStruct{},
			out: "Arguments:\n\t- Nested.Inner: integer (env: NESTED_INNER, flag: --nested.inner)\n",
		},
		{
			in:  &durationStruct{D: 5 * time.Hour},
			out: "Arguments:\n\t- D: duration (default: 5h0m0s) (env: D, flag: -d)\n",
		},
		{
			in:  &helpStructConfig{},
			out: "Arguments:\n\t- Yolo: string this is the help message (env: YOLO, flag: --yolo, -y)\n",
		},
	} {
		var (
			b bytes.Buffer

			cfg = NewDefaultConfigurator().(*helpConfigurator)
		)

		cfg.stderr = &b

		err := cfg.PrintDefaults(tt.in)

		assert.NoError(t, err)
		assert.Equal(t, tt.out, b.String())
	}
}
