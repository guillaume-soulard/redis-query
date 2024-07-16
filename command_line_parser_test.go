package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseArguments(t *testing.T) {
	tests := []struct {
		args           string
		expectedResult []string
	}{
		{args: `test`, expectedResult: []string{"test"}},
		{args: `1 2`, expectedResult: []string{"1", "2"}},
		{args: `"un test"`, expectedResult: []string{"un test"}},
		{args: `1 2 "un test"`, expectedResult: []string{"1", "2", "un test"}},
		{args: "1 2 \"un test\"", expectedResult: []string{"1", "2", "un test"}},
	}
	for _, test := range tests {
		name := fmt.Sprintf("ParseArguments should return %v when args is %s", test.expectedResult, test.args)
		t.Run(name, func(t *testing.T) {
			// GIVEN
			// WHEN
			result := ParseArguments(test.args)
			// THEN
			assert.Equal(t, test.expectedResult, result)
		})
	}
}
