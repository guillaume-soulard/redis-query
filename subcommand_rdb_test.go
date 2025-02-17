package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_bitsToByte(t *testing.T) {
	tests := []struct {
		bits          []byte
		expectedBytes []byte
	}{
		{
			bits:          []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expectedBytes: []byte{0},
		},
		{
			bits:          []byte{0, 0, 1, 1, 0, 0, 0, 0},
			expectedBytes: []byte{48},
		},
		{
			bits:          []byte{1, 1, 1, 1, 1, 1, 1, 1},
			expectedBytes: []byte{255},
		},
		{
			bits: []byte{
				0, 1, 1, 1, 0, 1, 0, 0,
				0, 1, 1, 0, 0, 1, 0, 1,
				0, 1, 1, 1, 0, 0, 1, 1,
				0, 1, 1, 1, 0, 1, 0, 0,
			},
			expectedBytes: []byte{116, 101, 115, 116},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("bitsToByte should_return %v when bits %v are passed", test.expectedBytes, test.bits)
		t.Run(name, func(t *testing.T) {
			// GIVEN
			// WHEN
			result := bitsToByte(test.bits)
			// THEN
			assert.Equal(t, test.expectedBytes, result)
		})
	}
}

func Test_bitsToInteger(t *testing.T) {
	tests := []struct {
		bits     []byte
		expected int64
	}{
		{
			bits:     []byte{0, 0, 0, 0, 0, 0},
			expected: 0,
		},
		{
			bits:     []byte{0, 0, 0, 0, 0, 1},
			expected: 1,
		},
		{
			bits:     []byte{0, 0, 0, 0, 1, 0},
			expected: 2,
		},
		{
			bits: []byte{
				0, 0, 0, 0, 0, 0, 1, 0,
				0, 0, 0, 0, 0, 0, 1, 0,
			},
			expected: 514,
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("bitsToInteger should_return %v when bits %v are passed", test.expected, test.bits)
		t.Run(name, func(t *testing.T) {
			// GIVEN
			// WHEN
			result := bitsToInteger(test.bits)
			// THEN
			assert.Equal(t, test.expected, result)
		})
	}
}
