package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"strings"
)

// ReadHost reads the host from CloudStack by ID.
func (client *Client) ReadHost(id string) (*models.HostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read host %s. details: %w", id, err)
	}

	host, _, err := client.CsClient.Host.GetHostByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	if host == nil {
		return nil, nil
	}

	return models.CreateHostPublicFromGet(host), nil
}

// ReadHostByName reads the host from CloudStack by name.
func (client *Client) ReadHostByName(name string) (*models.HostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read host %s. details: %w", name, err)
	}

	host, _, err := client.CsClient.Host.GetHostByName(name)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	if host == nil {
		return nil, nil
	}

	return models.CreateHostPublicFromGet(host), nil
}

// ReadHostByVM reads the host from CloudStack by VM ID.
func (client *Client) ReadHostByVM(vmID string) (*models.HostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read host for vm %s. details: %w", vmID, err)
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(vmID)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	if vm == nil {
		return nil, nil
	}

	if vm.Hostid == "" {
		return nil, nil
	}

	return client.ReadHost(vm.Hostid)
}
