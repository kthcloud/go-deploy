package log

import (
	"sync"
	"testing"

	"github.com/kthcloud/go-deploy/models/mode"
)

func TestDefaultLoggerIsNonNil(t *testing.T) {
	if Logger == nil {
		t.Fatal("expected default Logger to be non-nil")
	}
	// Should not panic even before SetupLogger is called.
	Debugln("nop debug")
	Infoln("nop info")
}

func TestNewLogger(t *testing.T) {
	for _, m := range []string{mode.Dev, mode.Test, mode.Prod} {
		l, err := New(m)
		if err != nil {
			t.Fatalf("New(%q) error = %v", m, err)
		}
		if l == nil {
			t.Fatalf("New(%q) returned nil logger", m)
		}
	}
}

func TestGetConcurrentSafe(t *testing.T) {
	if err := SetupLogger(mode.Test); err != nil {
		t.Fatalf("SetupLogger error = %v", err)
	}

	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := "worker"
			if id%2 == 0 {
				name = "other"
			}
			l := Get(name)
			if l == nil {
				t.Errorf("Get(%q) returned nil", name)
			}
		}(i)
	}
	wg.Wait()
}
