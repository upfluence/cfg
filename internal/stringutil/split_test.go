package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	for _, tt := range []struct {
		inStr  string
		inRune rune

		wantRes []string
		wantErr error
	}{
		{
			inStr:   "foo,bar",
			inRune:  ',',
			wantRes: []string{"foo", "bar"},
		},
		{
			inStr:   "'foo,bar',bar",
			inRune:  ',',
			wantRes: []string{"foo,bar", "bar"},
		},
		{
			inStr:   "'foo,\"bar,biz\"',bar",
			inRune:  ',',
			wantRes: []string{"foo,\"bar,biz\"", "bar"},
		},
		{
			inStr:   ",",
			inRune:  ',',
			wantRes: nil,
		},
		{
			inStr:   "foo,,",
			inRune:  ',',
			wantRes: []string{"foo"},
		},
		{
			inStr:   "foo,\"fuz",
			inRune:  ',',
			wantErr: errNotValid,
		},
		{
			inStr:   "foo='nested=k,v=z'",
			inRune:  ',',
			wantRes: []string{"foo='nested=k,v=z'"},
		},
	} {
		got, err := Split(tt.inStr, tt.inRune)

		assert.Equal(t, tt.wantErr, err)
		assert.Equal(t, tt.wantRes, got)
	}
}
