package table

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/x/cli"
)

type testRow struct {
	Name string
	Age  int
	City string
}

func testCommandContext(buf *bytes.Buffer) cli.CommandContext {
	return cli.CommandContext{
		Stdout:       buf,
		Configurator: cfg.NewDefaultConfigurator(),
	}
}

func TestNewPrinter(t *testing.T) {
	extFn := func(r testRow, col string) string {
		switch col {
		case "NAME":
			return r.Name
		case "AGE":
			return fmt.Sprintf("%d", r.Age)
		case "CITY":
			return r.City
		default:
			return ""
		}
	}

	for _, tc := range []struct {
		name      string
		haveKey   string
		haveFF    FormatterFunc
		haveRows  []testRow
		haveCols  []string
		haveExtFn func(testRow, string) string
		want      string
	}{
		{
			name:      "table/single row",
			haveKey:   "table",
			haveFF:    NewTabwriterFormatter,
			haveCols:  []string{"NAME", "AGE"},
			haveExtFn: extFn,
			haveRows:  []testRow{{Name: "alice", Age: 30}},
			want:      "NAME   AGE\nalice  30\n",
		},
		{
			name:      "table/multiple rows",
			haveKey:   "table",
			haveFF:    NewTabwriterFormatter,
			haveCols:  []string{"NAME", "CITY"},
			haveExtFn: extFn,
			haveRows: []testRow{
				{Name: "alice", City: "paris"},
				{Name: "bob", City: "london"},
			},
			want: "NAME   CITY\nalice  paris\nbob    london\n",
		},
		{
			name:      "table/empty slice",
			haveKey:   "table",
			haveFF:    NewTabwriterFormatter,
			haveCols:  []string{"NAME"},
			haveExtFn: func(_ testRow, _ string) string { return "" },
			haveRows:  []testRow{},
			want:      "NAME\n",
		},
		{
			name:      "csv/single row",
			haveKey:   "csv",
			haveFF:    NewCSVFormatter,
			haveCols:  []string{"NAME", "AGE"},
			haveExtFn: extFn,
			haveRows:  []testRow{{Name: "alice", Age: 30}},
			want:      "NAME,AGE\nalice,30\n",
		},
		{
			name:      "csv/multiple rows",
			haveKey:   "csv",
			haveFF:    NewCSVFormatter,
			haveCols:  []string{"NAME", "CITY"},
			haveExtFn: extFn,
			haveRows: []testRow{
				{Name: "alice", City: "paris"},
				{Name: "bob", City: "london"},
			},
			want: "NAME,CITY\nalice,paris\nbob,london\n",
		},
		{
			name:      "csv/empty slice",
			haveKey:   "csv",
			haveFF:    NewCSVFormatter,
			haveCols:  []string{"NAME"},
			haveExtFn: func(_ testRow, _ string) string { return "" },
			haveRows:  []testRow{},
			want:      "NAME\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPrinter[testRow](tc.haveKey, tc.haveFF, tc.haveCols, tc.haveExtFn)

			assert.Equal(t, tc.haveKey, p.Key())

			var buf bytes.Buffer

			err := p.Print(context.Background(), testCommandContext(&buf), tc.haveRows)

			require.NoError(t, err)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestNewDefaultPrinter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		haveKey  string
		haveFF   FormatterFunc
		haveRows []testRow
		want     string
	}{
		{
			name:     "table/single row",
			haveKey:  "table",
			haveFF:   NewTabwriterFormatter,
			haveRows: []testRow{{Name: "alice", Age: 30, City: "paris"}},
			want:     "Name   Age  City\nalice  30   paris\n",
		},
		{
			name:    "table/multiple rows",
			haveKey: "table",
			haveFF:  NewTabwriterFormatter,
			haveRows: []testRow{
				{Name: "alice", Age: 30, City: "paris"},
				{Name: "bob", Age: 25, City: "london"},
			},
			want: "Name   Age  City\nalice  30   paris\nbob    25   london\n",
		},
		{
			name:     "table/empty slice",
			haveKey:  "table",
			haveFF:   NewTabwriterFormatter,
			haveRows: []testRow{},
			want:     "Name  Age  City\n",
		},
		{
			name:     "csv/single row",
			haveKey:  "csv",
			haveFF:   NewCSVFormatter,
			haveRows: []testRow{{Name: "alice", Age: 30, City: "paris"}},
			want:     "Name,Age,City\nalice,30,paris\n",
		},
		{
			name:    "csv/multiple rows",
			haveKey: "csv",
			haveFF:  NewCSVFormatter,
			haveRows: []testRow{
				{Name: "alice", Age: 30, City: "paris"},
				{Name: "bob", Age: 25, City: "london"},
			},
			want: "Name,Age,City\nalice,30,paris\nbob,25,london\n",
		},
		{
			name:     "csv/empty slice",
			haveKey:  "csv",
			haveFF:   NewCSVFormatter,
			haveRows: []testRow{},
			want:     "Name,Age,City\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewDefaultPrinter[testRow](tc.haveKey, tc.haveFF)

			assert.Equal(t, tc.haveKey, p.Key())

			var buf bytes.Buffer

			err := p.Print(context.Background(), testCommandContext(&buf), tc.haveRows)

			require.NoError(t, err)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

type nestedInner struct {
	Value string
}

type nestedRow struct {
	ID    int
	Inner nestedInner
}

func TestNewDefaultPrinterNested(t *testing.T) {
	for _, tc := range []struct {
		name     string
		haveKey  string
		haveFF   FormatterFunc
		haveRows []nestedRow
		want     string
	}{
		{
			name:    "table",
			haveKey: "table",
			haveFF:  NewTabwriterFormatter,
			haveRows: []nestedRow{
				{ID: 1, Inner: nestedInner{Value: "foo"}},
				{ID: 2, Inner: nestedInner{Value: "bar"}},
			},
			want: "ID  Inner.Value\n1   foo\n2   bar\n",
		},
		{
			name:    "csv",
			haveKey: "csv",
			haveFF:  NewCSVFormatter,
			haveRows: []nestedRow{
				{ID: 1, Inner: nestedInner{Value: "foo"}},
				{ID: 2, Inner: nestedInner{Value: "bar"}},
			},
			want: "ID,Inner.Value\n1,foo\n2,bar\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewDefaultPrinter[nestedRow](tc.haveKey, tc.haveFF)

			var buf bytes.Buffer

			err := p.Print(context.Background(), testCommandContext(&buf), tc.haveRows)

			require.NoError(t, err)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

type taggedRow struct {
	Name  string `table:"name"`
	Email string `table:"email"`
	Age   int    `table:"-"`
}

func TestNewDefaultPrinterWithTableTags(t *testing.T) {
	p := NewDefaultPrinter[taggedRow]("table", NewTabwriterFormatter)

	var buf bytes.Buffer

	err := p.Print(context.Background(), testCommandContext(&buf), []taggedRow{
		{Name: "alice", Email: "alice@example.com", Age: 30},
	})

	require.NoError(t, err)
	assert.Equal(t, "name   email\nalice  alice@example.com\n", buf.String())
}

type csvTaggedRow struct {
	Name  string `csv:"name"`
	Email string `csv:"email"`
	Age   int    `csv:"-"`
}

func TestNewDefaultPrinterWithCSVTags(t *testing.T) {
	p := NewDefaultPrinter[csvTaggedRow]("csv", NewCSVFormatter)

	var buf bytes.Buffer

	err := p.Print(context.Background(), testCommandContext(&buf), []csvTaggedRow{
		{Name: "alice", Email: "alice@example.com", Age: 30},
	})

	require.NoError(t, err)
	assert.Equal(t, "name,email\nalice,alice@example.com\n", buf.String())
}
