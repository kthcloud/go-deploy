package system_state_poll

import (
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/host_repo"
	"go-deploy/pkg/db/resources/system_status_repo"
	wErrors "go-deploy/pkg/services/errors"
	"go-deploy/pkg/subsystems/host_api"
	"go-deploy/utils"
	"sync"
	"time"
)

func GetHostStatuses() ([]body.HostStatus, error) {
	allHosts, err := host_repo.New().Activated().List()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hosts. details: %s", err)
	}

	outputs := make([]*body.HostStatus, len(allHosts))
	mu := sync.Mutex{}

	err = ForEachHost("fetch-status", allHosts, func(idx int, host *model.Host) error {
		makeError := func(err error) error {
			return fmt.Errorf("failed to get status for host %s. details: %s", host.IP, err)
		}

		client := host_api.NewClient(host.ApiURL())

		status, err := client.GetStatus()
		if err != nil {
			return makeError(err)
		}

		hostStatus := body.HostStatus{
			Status: *status,
			HostBase: body.HostBase{
				Name:        host.Name,
				DisplayName: host.DisplayName,
				Zone:        host.Zone,
			},
		}

		mu.Lock()
		outputs[idx] = &hostStatus
		mu.Unlock()

		return nil
	})

	return utils.WithoutNils(outputs), err
}

func StatusWorker() error {
	hostStatuses, err := GetHostStatuses()
	if err != nil {
		return err
	}

	if len(hostStatuses) == 0 {
		return wErrors.NoHostsErr
	}

	status := body.SystemStatus{
		Hosts: hostStatuses,
	}

	return system_status_repo.New(500).Save(&body.TimestampedSystemStatus{
		Status:    status,
		Timestamp: time.Now(),
	})
}
