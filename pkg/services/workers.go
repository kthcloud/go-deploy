package services

import (
	"context"
	"fmt"
	"go-deploy/utils"
	"time"
)

// Worker is a wrapper for a worker main function and reports the workers status every second.
// It is meant to run asynchronously.
func Worker(ctx context.Context, name string, work func(context.Context) error) {
	defer OnStop(name)

	internalCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := work(internalCtx)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("%s failed: %w", name, err))
		}
	}()

	reportTick := time.Tick(1 * time.Second)
	cleanUpTick := time.Tick(300 * time.Second)

	for {
		select {
		case <-reportTick:
			ReportUp(name)
		case <-cleanUpTick:
			CleanUp()
		case <-internalCtx.Done():
			return
		case <-ctx.Done():
			return
		}
	}
}

// PeriodicWorker is a wrapper for a worker main function and calls it as often as the interval.
// It will never call a main function more than once at a time, so it might be slower than the interval
// if the main function takes longer than the interval to run.
// It is meant to run asynchronously.
func PeriodicWorker(ctx context.Context, name string, work func() error, interval time.Duration) {
	defer OnStop(name)

	reportTick := time.Tick(1 * time.Second)
	tick := time.Tick(interval)

	errorSleep := interval

	for {
		select {
		case <-reportTick:
			ReportUp(name)
		case <-tick:
			if err := work(); err != nil {
				utils.PrettyPrintError(fmt.Errorf("%s failed (sleeping for extra %s): %w", name, errorSleep.String(), err))
				time.Sleep(errorSleep)
				errorSleep *= 2
			} else {
				errorSleep = interval
			}
		case <-ctx.Done():
			return
		}
	}
}
