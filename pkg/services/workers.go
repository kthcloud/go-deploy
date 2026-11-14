package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kthcloud/go-deploy/pkg/db/resources/host_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	wErrors "github.com/kthcloud/go-deploy/pkg/services/errors"
	"github.com/kthcloud/go-deploy/utils"
)

type workerHooks struct {
	reportUp        func(string)
	cleanUp         func()
	onStop          func(string)
	deactivateHosts func([]string, time.Time) error
	onError         func(error)
	reportInterval  time.Duration
	cleanUpInterval time.Duration
}

func defaultWorkerHooks() workerHooks {
	return workerHooks{
		reportUp:        ReportUp,
		cleanUp:         CleanUp,
		onStop:          OnStop,
		deactivateHosts: DeactivateHosts,
		onError:         utils.PrettyPrintError,
		reportInterval:  1 * time.Second,
		cleanUpInterval: 300 * time.Second,
	}
}

// RunWorkers launches each worker function concurrently and waits until all of them return.
func RunWorkers(ctx context.Context, runners ...func(context.Context)) {
	var wg sync.WaitGroup
	for _, run := range runners {
		if run == nil {
			continue
		}
		wg.Add(1)
		go func(fn func(context.Context)) {
			defer wg.Done()
			fn(ctx)
		}(run)
	}
	wg.Wait()
}

// Worker is a wrapper for a worker main function and reports the worker's status every second.
// It cancels work when ctx is canceled and waits for work to return before exiting.
func Worker(ctx context.Context, name string, work func(context.Context) error) {
	runWorker(ctx, name, work, defaultWorkerHooks())
}

func runWorker(ctx context.Context, name string, work func(context.Context) error, h workerHooks) {
	defer h.onStop(name)

	internalCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := work(internalCtx); err != nil && !errors.Is(err, context.Canceled) {
			h.onError(fmt.Errorf("%s failed: %w", name, err))
			cancel()
		}
	}()

	reportTick := time.NewTicker(h.reportInterval)
	defer reportTick.Stop()

	cleanUpTick := time.NewTicker(h.cleanUpInterval)
	defer cleanUpTick.Stop()

	for {
		select {
		case <-reportTick.C:
			h.reportUp(name)
		case <-cleanUpTick.C:
			h.cleanUp()
		case <-done:
			return
		case <-internalCtx.Done():
			<-done
			return
		}
	}
}

// PeriodicWorker is a wrapper for a worker main function and calls it as often as the interval.
// It will never call a main function more than once at a time, so it might be slower than the interval
// if the main function takes longer than the interval to run.
func PeriodicWorker(ctx context.Context, name string, work func() error, interval time.Duration) {
	PeriodicWorkerCtx(ctx, name, func(context.Context) error { return work() }, interval)
}

// PeriodicWorkerCtx is like PeriodicWorker, passing ctx to work on each tick.
func PeriodicWorkerCtx(ctx context.Context, name string, work func(context.Context) error, interval time.Duration) {
	runPeriodicWorker(ctx, name, work, interval, defaultWorkerHooks())
}

func runPeriodicWorker(ctx context.Context, name string, work func(context.Context) error, interval time.Duration, h workerHooks) {
	defer h.onStop(name)

	if interval <= 0 {
		interval = time.Second
	}

	reportTick := time.NewTicker(h.reportInterval)
	defer reportTick.Stop()

	tick := time.NewTicker(interval)
	defer tick.Stop()

	errorSleep := interval

	for {
		select {
		case <-reportTick.C:
			h.reportUp(name)
		case <-tick.C:
			if err := work(ctx); err != nil {
				// If error is HostsFailedErr, disable the hosts.
				if hostsFailedErr, ok := errors.AsType[*wErrors.HostsFailedErr](err); ok {
					deactivateDuration := 30 * time.Minute
					log.Printf("Hosts [%s] failed when running %s. Deactivating them for %s", strings.Join(hostsFailedErr.Hosts, ", "), name, deactivateDuration.String())
					if deactivationErr := h.deactivateHosts(hostsFailedErr.Hosts, time.Now().Add(deactivateDuration)); deactivationErr != nil {
						h.onError(fmt.Errorf("failed to disable hosts: %w", deactivationErr))
					}
				}

				// It's too verbose to print when no hosts or clusters are available, so we skip that.
				if !errors.Is(err, wErrors.ErrNoClusters) && !errors.Is(err, wErrors.ErrNoHosts) {
					h.onError(fmt.Errorf("%s failed (sleeping for extra %s): %w", name, errorSleep.String(), err))
				}

				timer := time.NewTimer(errorSleep)
				select {
				case <-timer.C:
				case <-ctx.Done():
					timer.Stop()
					return
				}
				errorSleep *= 2
			} else {
				errorSleep = interval
			}
		case <-ctx.Done():
			return
		}
	}
}

// DeactivateHosts marks the given hosts as deactivated until the specified time.
func DeactivateHosts(hosts []string, until time.Time) error {
	for _, host := range hosts {
		err := host_repo.New().DeactivateHost(host, until)
		if err != nil {
			return fmt.Errorf("failed to deactivate host %s: %w", host, err)
		}
	}

	return nil
}
