package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	// GIVEN
	query := `{ PING }`
	// WHEN
	block, err := Parse(query)
	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, block)
}
