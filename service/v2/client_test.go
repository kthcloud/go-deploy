package v2

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewDefaultsNopLogger(t *testing.T) {
	c := New()
	if c.Logger() == nil {
		t.Fatal("expected non-nil default Logger()")
	}
}

func TestNewWithLoggerPreservesLogger(t *testing.T) {
	logger := zap.NewNop()
	c := NewWithLogger(logger)
	if c.Logger() != logger {
		t.Fatal("expected Logger() to return provided logger")
	}
}
