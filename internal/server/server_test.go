package server

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestStartWithInitErrorReturnsWrappedError(t *testing.T) {
	sentinel := errors.New("boom")
	workerRan := false
	shutdownCalled := false

	app, err := startWith(
		context.Background(),
		nil, // nil logger should default to zap.NewNop()
		[]initTask{
			{name: "ok", run: func(context.Context, *zap.Logger) error { return nil }},
			{name: "fail", run: func(context.Context, *zap.Logger) error { return sentinel }},
		},
		[]worker{
			{enabled: true, run: func(context.Context) { workerRan = true }},
		},
		nil,
		func(context.Context) error {
			shutdownCalled = true
			return nil
		},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected error wrapping %v, got %v", sentinel, err)
	}
	if app != nil {
		t.Fatalf("expected nil app on init failure")
	}
	if workerRan {
		t.Fatalf("expected workers not to start after init failure")
	}
	if shutdownCalled {
		t.Fatalf("expected shutdownFn not to be called after init failure")
	}
}

func TestStartWithWaitsForWorkersBeforeShutdown(t *testing.T) {
	var workerExited atomic.Bool
	var shutdownSawWorkerExited atomic.Bool
	workerStarted := make(chan struct{})

	app, err := startWith(
		context.Background(),
		nil,
		nil,
		[]worker{
			{
				enabled: true,
				run: func(ctx context.Context) {
					close(workerStarted)
					<-ctx.Done()
					time.Sleep(15 * time.Millisecond)
					workerExited.Store(true)
				},
			},
			{
				enabled: false,
				run: func(context.Context) {
					t.Error("disabled worker should not run")
				},
			},
		},
		nil,
		func(context.Context) error {
			shutdownSawWorkerExited.Store(workerExited.Load())
			return nil
		},
	)
	if err != nil {
		t.Fatalf("startWith error = %v", err)
	}

	<-workerStarted

	if err := app.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	// Stop should be idempotent.
	if err := app.Stop(); err != nil {
		t.Fatalf("second Stop() error = %v", err)
	}

	if !workerExited.Load() {
		t.Fatal("expected worker to exit on Stop()")
	}
	if !shutdownSawWorkerExited.Load() {
		t.Fatal("expected shutdownFn to run only after all workers exited")
	}
}

func TestValidateOptionsEnablesAllWhenNoneSelected(t *testing.T) {
	var opts Options
	if err := validateOptions(&opts, zap.NewNop()); err != nil {
		t.Fatalf("validateOptions error = %v", err)
	}
	if !opts.API || !opts.Confirmer || !opts.Cleaner || !opts.SystemStatePoller {
		t.Fatalf("expected all workers to be enabled when none specified, got %+v", opts)
	}
}
