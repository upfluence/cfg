package dflt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	p := Provider{}

	assert.Equal(t, "default", p.StructTag())

	for _, tc := range []struct {
		name    string
		haveKey string
		wantVal string
		wantOK  bool
		wantErr error
	}{
		{
			name:    "returns key as value",
			haveKey: "hello",
			wantVal: "hello",
			wantOK:  true,
		},
		{
			name:    "empty string skipped",
			haveKey: "",
			wantVal: "",
			wantOK:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			val, ok, err := p.Provide(context.Background(), tc.haveKey)

			require.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantVal, val)
		})
	}
}
