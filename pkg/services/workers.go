package services

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/pkg/db/resources/host_repo"
	"go-deploy/pkg/log"
	wErrors "go-deploy/pkg/services/errors"
	"go-deploy/utils"
	"strings"
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
			cancel()
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
				// If errors is HostsFailedErr, disable the hosts
				var hostsFailedErr *wErrors.HostsFailedErr
				if errors.As(err, &hostsFailedErr) {
					deactivateDuration := 30 * time.Minute
					log.Printf("Hosts [%s] failed. Deactivating them for %s", strings.Join(hostsFailedErr.Hosts, ", "), deactivateDuration.String())
					deactivationErr := DeactivateHosts(hostsFailedErr.Hosts, time.Now().Add(deactivateDuration))
					if deactivationErr != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to disable hosts: %w", deactivationErr))
					}
				}

				// It's too verbose to print when no hosts or clusters are available, so we skip that
				if !errors.Is(err, wErrors.NoClustersErr) && !errors.Is(err, wErrors.NoHostsErr) {
					utils.PrettyPrintError(fmt.Errorf("%s failed (sleeping for extra %s): %w", name, errorSleep.String(), err))
				}
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

func DeactivateHosts(hosts []string, until time.Time) error {
	for _, host := range hosts {
		err := host_repo.New().DeactivateHost(host, until)
		if err != nil {
			return fmt.Errorf("failed to deactivate host %s: %w", host, err)
		}
	}

	return nil
}
