package test

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

// TimeGTE checks if the first time is greater than the second time
func TimeGTE(t *testing.T, expected, actual time.Time, msgAndArgs ...interface{}) {
	t.Helper()

	assert.True(t, actual.After(expected) || actual.Equal(expected), msgAndArgs)
}

// TimeEq checks if two times are equal
func TimeEq(t *testing.T, expected, actual time.Time, msgAndArgs ...interface{}) {
	t.Helper()

	assert.True(t, expected.Equal(actual), msgAndArgs)
}

// TimeNotZero checks if the time is not zero
func TimeNotZero(t *testing.T, actual time.Time, msgAndArgs ...interface{}) {
	t.Helper()

	assert.NotEqual(t, time.Time{}, actual, msgAndArgs)
}

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
		assert.ElementsMatch(t, expected, actual, msgAndArgs)
	}
}

// NoError fails the test if there is an error
func NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if err != nil {
		assert.FailNow(t, err.Error(), msgAndArgs...)
	}
}
