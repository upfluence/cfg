package json

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_Provide(t *testing.T) {
	for _, tc := range []struct {
		name      string
		haveJSON  string
		haveKey   string
		wantValue string
		assertVal func(*testing.T, string)
		wantExist bool
		wantErr   error
	}{
		{
			name:     "empty store",
			haveJSON: `{}`,
			haveKey:  "foo",
		},
		{
			name:      "top level value",
			haveJSON:  `{"foo":"bar"}`,
			haveKey:   "foo",
			wantValue: "bar",
			wantExist: true,
		},
		{
			name:      "slice value",
			haveJSON:  `{"foo":[1,2,3]}`,
			haveKey:   "foo",
			wantValue: "1,2,3",
			wantExist: true,
		},
		{
			name:     "map value",
			haveJSON: `{"foo":{"foo":1,"bar":2}}`,
			haveKey:  "foo",
			assertVal: func(t *testing.T, got string) {
				t.Helper()
				assert.Contains(t, []string{"foo=1,bar=2", "bar=2,foo=1"}, got)
			},
			wantExist: true,
		},
		{
			name:      "second level value",
			haveJSON:  `{"foo":{"fiz":"bar"}}`,
			haveKey:   "foo.fiz",
			wantValue: "bar",
			wantExist: true,
		},
		{
			name:     "wrong format",
			haveJSON: `{"foo":{"fiz":"bar"}}`,
			haveKey:  "foo.fiz.buz",
			wantErr:  ErrJSONMalformated,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProviderFromReader(strings.NewReader(tc.haveJSON))

			got, gotExist, err := p.Provide(context.Background(), tc.haveKey)

			require.ErrorIs(t, err, tc.wantErr)

			if tc.assertVal != nil {
				tc.assertVal(t, got)
			} else {
				assert.Equal(t, tc.wantValue, got)
			}

			assert.Equal(t, tc.wantExist, gotExist)
		})
	}
}

func TestProvider_SubKeys(t *testing.T) {
	for _, tc := range []struct {
		name     string
		haveJSON string
		haveKey  string
		want     []string
	}{
		{
			name:     "empty store",
			haveJSON: `{}`,
			haveKey:  "workers",
			want:     nil,
		},
		{
			name:     "top level map keys",
			haveJSON: `{"workers":{"0":{"host":"h0"},"1":{"host":"h1"}}}`,
			haveKey:  "workers",
			want:     []string{"0", "1"},
		},
		{
			name:     "nested prefix",
			haveJSON: `{"db":{"shards":{"primary":{"host":"h1"},"replica":{"host":"h2"}}}}`,
			haveKey:  "db.shards",
			want:     []string{"primary", "replica"},
		},
		{
			name:     "prefix points to non-map value",
			haveJSON: `{"workers":"not-a-map"}`,
			haveKey:  "workers",
			want:     nil,
		},
		{
			name:     "prefix not found",
			haveJSON: `{"foo":{"bar":"baz"}}`,
			haveKey:  "missing",
			want:     nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProviderFromReader(strings.NewReader(tc.haveJSON))

			got, err := p.(*Provider).SubKeys(context.Background(), tc.haveKey)

			require.NoError(t, err)

			sort.Strings(got)
			sort.Strings(tc.want)
			assert.Equal(t, tc.want, got)
		})
	}
}
