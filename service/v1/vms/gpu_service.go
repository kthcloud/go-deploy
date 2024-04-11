package vms

import (
	"errors"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_repo"
	sErrors "go-deploy/service/errors"
	utils2 "go-deploy/service/utils"
	"go-deploy/service/v1/vms/cs_service"
	"go-deploy/service/v1/vms/opts"
	"go-deploy/utils"
	"strings"
	"time"
)

// GetGPU gets a GPU
//
// It uses service.AuthInfo to only return the model the requesting user has access to
func (c *Client) GetGPU(id string, opts ...opts.GetGpuOpts) (*model.GPU, error) {
	o := utils2.GetFirstOrDefault(opts)

	var usePrivilegedGPUs bool
	if !c.V1.HasAuth() || c.V1.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
		usePrivilegedGPUs = true
	}

	var excludedGpus []string
	if !usePrivilegedGPUs {
		excludedGpus = config.Config.GPU.PrivilegedGPUs
	}

	gmc := gpu_repo.New().WithExclusion(config.Config.GPU.ExcludedHosts, excludedGpus)

	if o.Zone != nil {
		gmc.WithZone(*o.Zone)
	}

	gpu, err := gmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	if gpu == nil {
		return nil, nil
	}

	if o.AvailableGPUs {
		err = c.CheckGpuHardwareAvailable(id)
		if err != nil {
			switch {
			case errors.Is(err, sErrors.GpuNotFoundErr):
				return nil, nil
			case errors.Is(err, sErrors.HostNotAvailableErr):
				return nil, nil
			}

			utils.PrettyPrintError(fmt.Errorf("error checking if gpu is in use. details: %w", err))
			return nil, nil
		}
	}

	return gpu, nil
}

// GetGpuByVM gets a GPU attached to a VM
// If the VM does not have a GPU attached, it will return nil
func (c *Client) GetGpuByVM(vmID string) (*model.GPU, error) {
	return gpu_repo.New().WithVM(vmID).Get()
}

// ListGPUs lists GPUs
//
// It uses service.AuthInfo to only return the resources the requesting user has access to
func (c *Client) ListGPUs(opts ...opts.ListGpuOpts) ([]model.GPU, error) {
	o := utils2.GetFirstOrDefault(opts)

	excludedGPUs := config.Config.GPU.ExcludedGPUs

	if c.V1.Auth() != nil && !c.V1.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
		effectiveRole := c.V1.Auth().GetEffectiveRole()

		if !effectiveRole.Permissions.UsePrivilegedGPUs {
			excludedGPUs = append(excludedGPUs, config.Config.GPU.PrivilegedGPUs...)
		}
	}

	gmc := gpu_repo.New().WithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs)

	if o.Pagination != nil {
		gmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.Zone != nil {
		gmc.WithZone(*o.Zone)
	}

	if o.AvailableGPUs {
		gmc.OnlyAvailable()
		gpus, _ := gmc.List()
		availableGPUs := make([]model.GPU, 0)

		for _, gpu := range gpus {
			err := c.CheckGpuHardwareAvailable(gpu.ID)
			if err != nil {
				switch {
				case errors.Is(err, sErrors.GpuNotFoundErr):
					continue
				case errors.Is(err, sErrors.HostNotAvailableErr):
					continue
				default:
					utils.PrettyPrintError(fmt.Errorf("error checking if gpu is in use. details: %w", err))
					continue
				}
			}

			availableGPUs = append(availableGPUs, gpu)
		}

		return availableGPUs, nil
	}

	return gmc.List()
}

// AttachGPU attaches a GPU to a VM
func (c *Client) AttachGPU(vmID string, gpuIDs []string, leaseDuration float64) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu to vm %s. details: %w", vmID, err)
	}
	csInsufficientCapacityError := "host has capacity? false"
	gpuAlreadyAttachedError := "Unable to create a deployment for VM instance"

	var endLease time.Time
	if leaseDuration == -1 {
		// Represent "end" of time
		endLease = time.Now().AddDate(1000, 0, 0)
	} else {
		endLease = time.Now().Add(time.Duration(leaseDuration) * time.Hour)
	}

	vm, err := c.Get(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return sErrors.VmNotFoundErr
	}

	// Service client
	cc := cs_service.New(c.Cache)

	for _, gpuID := range gpuIDs {
		var gpu *model.GPU
		gpu, err = gpu_repo.New().GetByID(gpuID)
		if err != nil {
			return makeError(err)
		}

		if gpu == nil {
			continue
		}

		err = c.CheckGpuHardwareAvailable(gpuID)
		if err != nil {
			if errors.Is(err, sErrors.GpuNotFoundErr) {
				continue
			}

			return makeError(err)
		}

		if gpu.Lease.VmID != vmID && gpu.IsAttached() {
			// if it is attached but expired, take over the card by first detaching it
			if gpu.Lease.IsExpired() {
				err = cc.DetachGPU(gpu.Lease.VmID, cs_service.CsDetachGpuAfterStateRestore)
				if err != nil {
					return makeError(err)
				}

				err = gpu_repo.New().Detach(gpu.Lease.VmID)
				if err != nil {
					return makeError(err)
				}
			} else {
				continue
			}
		}

		var attached bool
		attached, err = gpu_repo.New().Attach(gpuID, vmID, vm.OwnerID, endLease)
		if err != nil {
			if errors.Is(err, gpu_repo.AlreadyAttachedErr) || errors.Is(err, gpu_repo.NotFoundErr) {
				// This is not treated as an error, just another instance snatched the GPU before this one
				continue
			}

			return makeError(err)
		}

		if !attached {
			// This is an edge case where we don't want to fail the method, since a retry will probably not help.
			//
			// This is probably caused by a race condition where two users requested the same GPU, where the first one
			// got it, and the second one failed. We don't want to fail the second user, since that would mean that a
			// job would get stuck.
			// Instead the user is not granted the GPU, and will need to request a new one manually
			continue
		}

		// Ensure it is attached in cloudstack, this will not do anything if it is already attached
		// otherwise, it will restart the vm, which is fine since the user requested this
		err = cc.AttachGPU(vmID, gpuID)
		if err == nil {
			break
		}

		errString := err.Error()

		insufficientCapacityErr := strings.Contains(errString, csInsufficientCapacityError)
		gpuAlreadyAttached := strings.Contains(errString, gpuAlreadyAttachedError)

		if insufficientCapacityErr {
			// If the host has insufficient capacity, we need to detach the GPU from the vm
			// and attempt to attach it to another GPU

			err = cc.DetachGPU(vmID, cs_service.CsDetachGpuAfterStateRestore)
			if err != nil {
				return makeError(err)
			}

			err = gpu_repo.New().Detach(vmID)
			if err != nil {
				return makeError(err)
			}
		} else if gpuAlreadyAttached {
			// If the GPU is already attached, we need to detach it from the vm

			err = cc.DetachGPU(vmID, cs_service.CsDetachGpuAfterStateRestore)
			if err != nil {
				return makeError(err)
			}

			err = gpu_repo.New().Detach(vmID)
			if err != nil {
				return makeError(err)
			}
		} else {
			return makeError(err)
		}
	}

	return nil
}

// DetachGPU detaches a GPU from a VM
func (c *Client) DetachGPU(vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from vm %s. details: %w", vmID, err)
	}

	err := cs_service.New(c.Cache).DetachGPU(vmID, cs_service.CsDetachGpuAfterStateRestore)
	if err != nil {
		return makeError(err)
	}

	err = gpu_repo.New().Detach(vmID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// IsGpuPrivileged checks if a GPU is privileged.
func (c *Client) IsGpuPrivileged(id string) (bool, error) {
	gpu, err := c.GPU(id, nil)
	if err != nil {
		return false, err
	}

	if gpu == nil {
		return false, nil
	}

	for _, privilegedGPU := range config.Config.GPU.PrivilegedGPUs {
		if privilegedGPU == gpu.Data.Name {
			return true, nil
		}
	}

	return false, nil
}

// CheckGpuHardwareAvailable checks if a GPU is available for use.
func (c *Client) CheckGpuHardwareAvailable(gpuID string) error {
	gpu, err := c.GPU(gpuID, nil)
	if err != nil {
		return err
	}

	if gpu == nil {
		return sErrors.GpuNotFoundErr
	}

	cc := cs_service.New(c.Cache)

	cloudstackAttached, err := cc.IsGpuAttached(gpuID)
	if err != nil {
		return err
	}

	zone := config.Config.VM.GetZone(gpu.Zone)
	if zone == nil {
		return sErrors.ZoneNotFoundErr
	}

	err = cc.CheckHostState(gpu.Host, zone)
	if err != nil {
		return err
	}

	// Check if it is a "bad attach", where cloudstack reports it being attached, but the database says it's not.
	// This usually means it is in use outside the scope of deploy
	if cloudstackAttached && !gpu.IsAttached() {
		return sErrors.GpuNotFoundErr
	}

	return nil
}

// CheckSuitableHost checks if a host is suitable for a VM.
// This is used to minimize the risk of starting a VM that cannot be started on a given host.
func (c *Client) CheckSuitableHost(id, hostName, zoneName string) error {
	vm, err := c.VM(id, nil)
	if err != nil {
		return err
	}

	if vm == nil {
		return sErrors.VmNotFoundErr
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return sErrors.VmNotCreatedErr
	}

	zone := config.Config.VM.GetZone(zoneName)
	if zone == nil {
		return sErrors.ZoneNotFoundErr
	}

	err = cs_service.New(c.Cache).CheckSuitableHost(vm.ID, hostName, zone)
	if err != nil {
		return err
	}

	return nil
}
