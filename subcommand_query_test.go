package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Parse_should_parse_ping(t *testing.T) {
	// GIVEN
	query := `{ PING }`
	// WHEN
	block, err := Parse(query)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "PING", *block.Block.Commands[0].Name)
}

func Test_Parse_should_parse_set_foo_bar(t *testing.T) {
	// GIVEN
	query := `{ SET 'foo' 'bar' }`
	// WHEN
	block, err := Parse(query)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "SET", *block.Block.Commands[0].Name)
	assert.Equal(t, "foo", *block.Block.Commands[0].Args[0].String)
	assert.Equal(t, "bar", *block.Block.Commands[0].Args[1].String)
}

func Test_Parse_should_parse_2_commands(t *testing.T) {
	// GIVEN
	query := `{ SET 'foo' 'bar' GET 'test' }`
	// WHEN
	block, err := Parse(query)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "SET", *block.Block.Commands[0].Name)
	assert.Equal(t, "foo", *block.Block.Commands[0].Args[0].String)
	assert.Equal(t, "bar", *block.Block.Commands[0].Args[1].String)
	assert.Equal(t, "GET", *block.Block.Commands[1].Name)
	assert.Equal(t, "test", *block.Block.Commands[1].Args[0].String)
}

func Test_Parse_should_parse_2_nested_commands(t *testing.T) {
	// GIVEN
	query := `{ GET 'test' -> { SET 'foo' 'bar' }}`
	// WHEN
	block, err := Parse(query)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "GET", *block.Block.Commands[0].Name)
	assert.Equal(t, "test", *block.Block.Commands[0].Args[0].String)
	assert.Equal(t, "SET", *block.Block.Commands[0].Block.Commands[0].Name)
	assert.Equal(t, "foo", *block.Block.Commands[0].Block.Commands[0].Args[0].String)
	assert.Equal(t, "bar", *block.Block.Commands[0].Block.Commands[0].Args[1].String)
}

func Test_Parse_should_parse_2_nested_commands_with_args(t *testing.T) {
	// GIVEN
	query := `{ GET 'test' -> param -> { SET 'foo' $param }}`
	// WHEN
	block, err := Parse(query)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "GET", *block.Block.Commands[0].Name)
	assert.Equal(t, "test", *block.Block.Commands[0].Args[0].String)
	assert.Equal(t, "param", block.Block.Commands[0].Block.BlockArgs.Args[0])
	assert.Equal(t, "SET", *block.Block.Commands[0].Block.Commands[0].Name)
	assert.Equal(t, "foo", *block.Block.Commands[0].Block.Commands[0].Args[0].String)
	assert.Equal(t, "param", *block.Block.Commands[0].Block.Commands[0].Args[1].Variable)
}
