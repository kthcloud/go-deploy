package services

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func nopHooks() workerHooks {
	return workerHooks{
		reportUp:        func(string) {},
		cleanUp:         func() {},
		onStop:          func(string) {},
		deactivateHosts: func([]string, time.Time) error { return nil },
		onError:         func(error) {},
		reportInterval:  50 * time.Millisecond,
		cleanUpInterval: time.Second,
	}
}

func TestRunWorkersWaitsForAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var finished atomic.Int32

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunWorkers(ctx,
			func(ctx context.Context) {
				<-ctx.Done()
				time.Sleep(10 * time.Millisecond)
				finished.Add(1)
			},
			func(ctx context.Context) {
				<-ctx.Done()
				finished.Add(1)
			},
			nil,
		)
	}()

	cancel()
	<-done

	if got := finished.Load(); got != 2 {
		t.Fatalf("expected 2 workers to finish, got %d", got)
	}
}

func TestWorkerCancelsAndWaitsForWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var workExited atomic.Bool
	var stopped atomic.Bool

	h := nopHooks()
	h.onStop = func(string) {
		if !workExited.Load() {
			t.Error("onStop called before work exited")
		}
		stopped.Store(true)
	}

	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		runWorker(ctx, "test-worker", func(wCtx context.Context) error {
			close(started)
			<-wCtx.Done()
			time.Sleep(15 * time.Millisecond)
			workExited.Store(true)
			return wCtx.Err()
		}, h)
	}()

	<-started
	cancel()
	<-done

	if !workExited.Load() || !stopped.Load() {
		t.Fatal("expected work and onStop to complete")
	}
}

func TestPeriodicWorkerExitsPromptlyDuringErrorBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	h := nopHooks()

	errored := make(chan struct{}, 1)
	h.onError = func(error) {
		select {
		case errored <- struct{}{}:
		default:
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Long error backoff (10s) must be interrupted immediately on cancel.
		runPeriodicWorker(ctx, "test-periodic", func(context.Context) error {
			return errors.New("transient error")
		}, 10*time.Millisecond, h)
	}()

	<-errored
	start := time.Now()
	cancel()

	select {
	case <-done:
		if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
			t.Fatalf("expected prompt exit on cancel, took %s", elapsed)
		}
	case <-time.After(time.Second):
		t.Fatal("periodic worker did not exit promptly during error backoff")
	}
}
