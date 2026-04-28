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

type helpString string

func (h helpString) Help() string { return string(h) }

type helperFieldConfig struct {
	Dynamic helpString `env:"-" flag:"dyn"`
}

type helperOverridesTagConfig struct {
	Dynamic helpString `env:"-" flag:"dyn" help:"from tag"`
}

type helperEmptyFallsBackConfig struct {
	Dynamic helpString `env:"-" flag:"dyn" help:"from tag"`
}

type mapStructConfig struct {
	Databases map[string]dbConfig `env:"DATABASES" flag:"databases"`
}

type sliceStructConfig struct {
	Workers []dbConfig `env:"WORKERS" flag:"workers"`
}

type mapPtrStructConfig struct {
	Databases map[string]*dbConfig `env:"DATABASES" flag:"databases"`
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
		{
			name: "Help() method provides help text",
			in:   &helperFieldConfig{Dynamic: "dynamic help"},
			out: "Arguments:\n" +
				"\t- Dynamic: string dynamic help (default: dynamic help) (flag: --dyn)\n",
		},
		{
			name: "Help() method overrides struct tag",
			in:   &helperOverridesTagConfig{Dynamic: "from method"},
			out: "Arguments:\n" +
				"\t- Dynamic: string from method (default: from method) (flag: --dyn)\n",
		},
		{
			name: "empty Help() falls back to struct tag",
			in:   &helperEmptyFallsBackConfig{},
			out: "Arguments:\n" +
				"\t- Dynamic: string from tag (flag: --dyn)\n",
		},
		{
			name: "map of structs shows inner fields with <key> placeholder",
			in:   &mapStructConfig{},
			out: "Arguments:\n" +
				"\t- Databases.<key>.Host: string (env: DATABASES_<KEY>_HOST, flag: --databases.<key>.host)\n" +
				"\t- Databases.<key>.Port: integer (env: DATABASES_<KEY>_PORT, flag: --databases.<key>.port)\n",
		},
		{
			name: "slice of structs shows inner fields with <N> placeholder",
			in:   &sliceStructConfig{},
			out: "Arguments:\n" +
				"\t- Workers.<N>.Host: string (env: WORKERS_<N>_HOST, flag: --workers.<n>.host)\n" +
				"\t- Workers.<N>.Port: integer (env: WORKERS_<N>_PORT, flag: --workers.<n>.port)\n",
		},
		{
			name: "map of ptr structs shows inner fields with <key> placeholder",
			in:   &mapPtrStructConfig{},
			out: "Arguments:\n" +
				"\t- Databases.<key>.Host: string (env: DATABASES_<KEY>_HOST, flag: --databases.<key>.host)\n" +
				"\t- Databases.<key>.Port: integer (env: DATABASES_<KEY>_PORT, flag: --databases.<key>.port)\n",
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
