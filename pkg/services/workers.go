package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kthcloud/go-deploy/pkg/db/resources/host_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	wErrors "github.com/kthcloud/go-deploy/pkg/services/errors"
	"github.com/kthcloud/go-deploy/utils"
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

	reportTick := time.NewTicker(1 * time.Second)
	cleanUpTick := time.NewTicker(300 * time.Second)

	for {
		select {
		case <-reportTick.C:
			ReportUp(name)
		case <-cleanUpTick.C:
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

	reportTick := time.NewTicker(1 * time.Second)
	tick := time.NewTicker(interval)

	errorSleep := interval

	for {
		select {
		case <-reportTick.C:
			ReportUp(name)
		case <-tick.C:
			if err := work(); err != nil {
				// If errors is HostsFailedErr, disable the hosts
				var hostsFailedErr *wErrors.HostsFailedErr
				if errors.As(err, &hostsFailedErr) {
					deactivateDuration := 30 * time.Minute
					log.Printf("Hosts [%s] failed when running %s. Deactivating them for %s", strings.Join(hostsFailedErr.Hosts, ", "), name, deactivateDuration.String())
					deactivationErr := DeactivateHosts(hostsFailedErr.Hosts, time.Now().Add(deactivateDuration))
					if deactivationErr != nil {
						utils.PrettyPrintError(fmt.Errorf("failed to disable hosts: %w", deactivationErr))
					}
				}

				// It's too verbose to print when no hosts or clusters are available, so we skip that
				if !errors.Is(err, wErrors.ErrNoClusters) && !errors.Is(err, wErrors.ErrNoHosts) {
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
