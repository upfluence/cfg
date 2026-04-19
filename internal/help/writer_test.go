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
	Sub    mapStringIntStruct `env:"-" flag:"-"`
	Parser parserImpl
}

type parserImpl struct {
	Foo string
}

func (*parserImpl) Parse(string) error { return nil }

type defaultTagStruct struct {
	Host string `default:"localhost" env:"HOST" flag:"host"`
	Port int    `env:"PORT" flag:"port"`
}

type nestedDefaultStruct struct {
	DB dbConfig `env:"DB" flag:"db"`
}

type dbConfig struct {
	Host string `default:"localhost" env:"HOST" flag:"host"`
	Port int    `default:"5432" env:"PORT" flag:"port"`
}

func TestPrintDefaults(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   interface{}
		out  string
	}{
		{
			name: "map with default value",
			in:   &mapStringIntStruct{Map: map[string]int{"fiz": 42}},
			out: "Arguments:\n" +
				"\t- Map: map[string]integer (default: map[fiz:42]) (env: MAP, flag: --map)\n",
		},
		{
			name: "help message",
			in:   &helpStructConfig{},
			out: "Arguments:\n" +
				"\t- Yolo: string this is the help message (flag: --yolo, -y)\n",
		},
		{
			name: "nested struct",
			in:   &nestedStruct{},
			out: "Arguments:\n" +
				"\t- Parser: help.parserImpl (env: PARSER, flag: --parser)\n",
		},
		{
			name: "default tag",
			in:   &defaultTagStruct{},
			out: "Arguments:\n" +
				"\t- Host: string (default: localhost) (env: HOST, flag: --host)\n" +
				"\t- Port: integer (env: PORT, flag: --port)\n",
		},
		{
			name: "default tag with pre-existing value",
			in:   &defaultTagStruct{Host: "example.com", Port: 8080},
			out: "Arguments:\n" +
				"\t- Host: string (default: localhost) (env: HOST, flag: --host)\n" +
				"\t- Port: integer (default: 8080) (env: PORT, flag: --port)\n",
		},
		{
			name: "nested struct with default tags",
			in:   &nestedDefaultStruct{},
			out: "Arguments:\n" +
				"\t- DB.Host: string (default: localhost) (env: DB_HOST, flag: --db.host)\n" +
				"\t- DB.Port: integer (default: 5432) (env: DB_PORT, flag: --db.port)\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer

			_, err := DefaultWriter.Write(&b, tt.in)

			assert.NoError(t, err)
			assert.Equal(t, tt.out, b.String())
		})
	}
}
