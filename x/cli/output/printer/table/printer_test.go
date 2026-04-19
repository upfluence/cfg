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
	for _, tc := range []struct {
		name      string
		haveRows  []testRow
		haveCols  []string
		haveExtFn func(testRow, string) string
		want      string
	}{
		{
			name:     "single row",
			haveCols: []string{"NAME", "AGE"},
			haveExtFn: func(r testRow, col string) string {
				switch col {
				case "NAME":
					return r.Name
				case "AGE":
					return fmt.Sprintf("%d", r.Age)
				default:
					return ""
				}
			},
			haveRows: []testRow{{Name: "alice", Age: 30}},
			want:     "NAME   AGE\nalice  30\n",
		},
		{
			name:     "multiple rows",
			haveCols: []string{"NAME", "CITY"},
			haveExtFn: func(r testRow, col string) string {
				switch col {
				case "NAME":
					return r.Name
				case "CITY":
					return r.City
				default:
					return ""
				}
			},
			haveRows: []testRow{
				{Name: "alice", City: "paris"},
				{Name: "bob", City: "london"},
			},
			want: "NAME   CITY\nalice  paris\nbob    london\n",
		},
		{
			name:      "empty slice",
			haveCols:  []string{"NAME"},
			haveExtFn: func(_ testRow, _ string) string { return "" },
			haveRows:  []testRow{},
			want:      "NAME\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPrinter[testRow](tc.haveCols, tc.haveExtFn)

			assert.Equal(t, "table", p.Key())

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
		haveRows []testRow
		want     string
	}{
		{
			name:     "single row",
			haveRows: []testRow{{Name: "alice", Age: 30, City: "paris"}},
			want:     "Name   Age  City\nalice  30   paris\n",
		},
		{
			name: "multiple rows",
			haveRows: []testRow{
				{Name: "alice", Age: 30, City: "paris"},
				{Name: "bob", Age: 25, City: "london"},
			},
			want: "Name   Age  City\nalice  30   paris\nbob    25   london\n",
		},
		{
			name:     "empty slice",
			haveRows: []testRow{},
			want:     "Name  Age  City\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewDefaultPrinter[testRow]()

			assert.Equal(t, "table", p.Key())

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
	p := NewDefaultPrinter[nestedRow]()

	var buf bytes.Buffer

	err := p.Print(context.Background(), testCommandContext(&buf), []nestedRow{
		{ID: 1, Inner: nestedInner{Value: "foo"}},
		{ID: 2, Inner: nestedInner{Value: "bar"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "ID  Inner.Value\n1   foo\n2   bar\n", buf.String())
}

type taggedRow struct {
	Name  string `table:"name"`
	Email string `table:"email"`
	Age   int    `table:"-"`
}

func TestNewDefaultPrinterWithTableTags(t *testing.T) {
	p := NewDefaultPrinter[taggedRow]()

	var buf bytes.Buffer

	err := p.Print(context.Background(), testCommandContext(&buf), []taggedRow{
		{Name: "alice", Email: "alice@example.com", Age: 30},
	})

	require.NoError(t, err)
	assert.Equal(t, "name   email\nalice  alice@example.com\n", buf.String())
}
