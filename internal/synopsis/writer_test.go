package synopsis

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type helpStructConfig struct {
	Yolo string `flag:"yolo,y"`
	Bar  string
}

type mapStringIntStruct struct {
	Map map[string]int `flag:"map"`
}

type nestedStruct struct {
	Sub    mapStringIntStruct `flag:"-"`
	Parser *parserImpl
}

type parserImpl struct {
	Foo string
}

func (*parserImpl) Parse(string) error { return nil }

func TestPrintDefaults(t *testing.T) {
	for _, tt := range []struct {
		in  interface{}
		out string
	}{
		{
			in:  &mapStringIntStruct{Map: map[string]int{"fiz": 42}},
			out: "[--map] ",
		},
		{
			in:  &helpStructConfig{},
			out: "[--yolo, -y] [--bar] ",
		},
		{
			in:  &nestedStruct{},
			out: "[--parser] ",
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
