package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var t0 time.Time

	err := Parse("1970-01-02", &t0, WithDateFormat("2006-01-02"))
	assert.NoError(t, err)
	assert.Equal(t, int64(86400), t0.Unix())

	var vs []int64

	err = Parse("1,2,3", &vs)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, vs)
}
