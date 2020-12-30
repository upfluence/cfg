package help

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type helpStructConfig struct {
	Yolo string `help:"this is the help message" flag:"yolo,y" env:"-"`
}

type mapStringIntStruct struct {
	Map map[string]int `mock:"map"`
}

type nestedStruct struct {
	Sub mapStringIntStruct `env:"-" flag:"-"`
}

func TestPrintDefaults(t *testing.T) {
	for _, tt := range []struct {
		in  interface{}
		out string
	}{
		{
			in:  &mapStringIntStruct{Map: map[string]int{"fiz": 42}},
			out: "Arguments:\n\t- Map: map[string]integer (default: map[fiz:42]) (env: MAP, flag: --map)\n",
		},
		{
			in:  &helpStructConfig{},
			out: "Arguments:\n\t- Yolo: string this is the help message (flag: --yolo, -y)\n",
		},
		{
			in:  &nestedStruct{},
			out: "Arguments:\n",
		},
	} {
		var (
			b bytes.Buffer

			_, err = DefaultWriter.Write(&b, tt.in)
		)

		assert.NoError(t, err)
		assert.Equal(t, tt.out, b.String())
	}
}
