package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringInSlice(t *testing.T) {

	args := []string{"lorem", "ipsum", "dolor"}
	assert.True(t, StringInSlice("ipsum", args))
	assert.False(t, StringInSlice("sit", args))
}
