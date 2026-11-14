package key_value

import (
	"context"
	"testing"
)

func TestNewWithClientDefaultsNopLogger(t *testing.T) {
	c := NewWithClient(nil, nil)
	if c.Logger == nil {
		t.Fatal("expected Logger to default to zap.NewNop()")
	}
	if c.log() == nil {
		t.Fatal("expected log() helper to return non-nil logger")
	}
}

func TestSetUpExpirationListenerRejectsInvalidRegex(t *testing.T) {
	c := NewWithClient(nil, nil)
	err := c.SetUpExpirationListener(context.Background(), "[invalid", func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}
