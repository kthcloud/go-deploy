package test

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// EqualOrEmpty checks if two lists are equal, where [] == nil
func EqualOrEmpty(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// Check if "expected" is a slice, and if so, check how many elements it has
	isSlice := reflect.ValueOf(expected).Kind() == reflect.Slice
	noElements := 0
	if isSlice {
		noElements = reflect.ValueOf(expected).Len()
	}

	if expected == nil || (isSlice && noElements == 0) {
		assert.Empty(t, actual, msgAndArgs)
	} else {
		assert.EqualValues(t, expected, actual, msgAndArgs)
	}
}

// NoError fails the test if there is an error
func NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if err != nil {
		assert.FailNow(t, err.Error(), msgAndArgs...)
	}
}
