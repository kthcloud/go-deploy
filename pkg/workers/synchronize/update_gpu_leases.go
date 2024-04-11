package synchronize

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/gpu_group_repo"
	"go-deploy/pkg/db/resources/gpu_lease_repo"
	"go-deploy/pkg/log"
	"go-deploy/service"
	"sort"
	"time"
)

func gpuLeaseSynchronizer() error {
	gpuGroups, err := gpu_group_repo.New().List()
	if err != nil {
		return err
	}

	gpuLeases, err := gpu_lease_repo.New().List()
	if err != nil {
		return err
	}

	// Ensure that the GPU leases are sorted by createdAt to simplify queue position logic
	sort.Slice(gpuLeases, func(i, j int) bool { return gpuLeases[i].CreatedAt.Before(gpuLeases[j].CreatedAt) })

	// Check for leases that refer to non-existent GPU groups
	err = checkForNonExistentGpuGroups(gpuLeases, gpuGroups)
	if err != nil {
		return err
	}

	// Check for expired leases
	err = checkForExpiredLeases(gpuLeases)
	if err != nil {
		return err
	}

	// Check for leases that have a queue position of 0 without being assigned (They should be assigned)
	err = checkForUnassignedLeases(gpuLeases, gpuGroups)
	if err != nil {
		return err
	}

	// Check for leases that were assigned 24 hours ago and are not yet activated
	err = checkForUnactivatedLeases(gpuLeases)
	if err != nil {
		return err
	}

	// Check for leases that are expired in a queue with more leases than the total GPUs
	// If there are no pending leases, it can remain.
	err = checkForExpiredLeasesInFullQueues(gpuLeases, gpuGroups)
	if err != nil {
		return err
	}

	return nil
}

func checkForNonExistentGpuGroups(leases []model.GpuLease, groups []model.GpuGroup) error {
	groupIDs := make(map[string]bool)
	for _, group := range groups {
		groupIDs[group.ID] = true
	}

	deployV2 := service.V2()

	for _, lease := range leases {
		if _, ok := groupIDs[lease.GpuGroupID]; !ok {
			log.Infoln("Deleting lease", lease.ID, "since it refers to a non-existent GPU group")
			// The lease refers to a non-existent GPU group
			err := deployV2.VMs().GpuLeases().Delete(lease.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkForExpiredLeases(leases []model.GpuLease) error {
	now := time.Now()
	for _, lease := range leases {
		if lease.ActivatedAt == nil {
			// Lease is not activated yet, so it cannot be expired
			continue
		}

		if lease.ExpiredAt != nil {
			// Lease is already expired
			continue
		}

		expiresAt := lease.ActivatedAt.Add(time.Duration(lease.LeaseDuration) * time.Hour)
		if expiresAt.Before(now) {
			// Lease is expired, set the expiredAt field
			lease.ExpiredAt = &now
			err := gpu_lease_repo.New().MarkExpired(lease.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkForUnassignedLeases(leases []model.GpuLease, groups []model.GpuGroup) error {
	// Index the GPU leases by GPU ID
	leasesByGpuID := make(map[string][]model.GpuLease)
	for _, lease := range leases {
		leasesByGpuID[lease.GpuGroupID] = append(leasesByGpuID[lease.GpuGroupID], lease)
	}

	now := time.Now()
	for _, group := range groups {
		leases, ok := leasesByGpuID[group.ID]
		if !ok {
			continue
		}

		for _, lease := range leases {
			// Leases that are already assigned are guaranteed to have a queue position of 0
			if lease.AssignedAt != nil {
				continue
			}

			leasesInFront := 0
			for _, l := range leases {
				if l.CreatedAt.Before(lease.CreatedAt) {
					leasesInFront++
				}
			}

			queuePosition := max((leasesInFront-group.Total)+1, 0)
			if queuePosition == 0 {
				// Lease can be assigned
				lease.AssignedAt = &now
				err := gpu_lease_repo.New().MarkAssigned(lease.ID)
				if err != nil {
					return err
				}
			}
		}

		leasesByGpuID[group.ID] = leases
	}

	return nil
}

func checkForUnactivatedLeases(leases []model.GpuLease) error {
	now := time.Now()
	for _, lease := range leases {
		if lease.AssignedAt != nil && lease.ActivatedAt == nil && lease.AssignedAt.Before(now.Add(-24*time.Hour)) {
			// Lease is assigned 24 hours ago and is not yet activated by the user
			lease.ActivatedAt = &now
			err := gpu_lease_repo.New().MarkActivated(lease.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkForExpiredLeasesInFullQueues(leases []model.GpuLease, groups []model.GpuGroup) error {
	// Index the GPU leases by GPU ID
	leasesByGpuID := make(map[string][]model.GpuLease)
	for _, lease := range leases {
		leasesByGpuID[lease.GpuGroupID] = append(leasesByGpuID[lease.GpuGroupID], lease)
	}

	deployV2 := service.V2()

	for _, group := range groups {
		leasesByGPU, ok := leasesByGpuID[group.ID]
		if !ok {
			continue
		}

		pending := max(len(leasesByGPU)-group.Total, 0)
		if pending == 0 {
			// The group is not full, any expired leases can remain
			continue
		}

		// The group has pending leases, check for expired leases and delete them
		// Delete only the amount needed, and sort by expiration date to delete the ones that have been expired the longest
		leasesByExpiration := make([]model.GpuLease, len(leasesByGPU))
		copy(leasesByExpiration, leasesByGPU)

		// If expiredAt is null, it should be last in the list
		sort.Slice(leasesByExpiration, func(i, j int) bool {
			if leasesByExpiration[i].ExpiredAt == nil {
				return false
			} else if leasesByExpiration[j].ExpiredAt == nil {
				return true
			} else {
				return leasesByExpiration[i].ExpiredAt.Before(*leasesByExpiration[j].ExpiredAt)
			}
		})

		noDeleted := 0
		for _, lease := range leasesByExpiration {
			if !lease.IsExpired() {
				// No more leases are expired since the list is sorted
				break
			}

			err := deployV2.VMs().GpuLeases().Delete(lease.ID)
			if err != nil {
				return err
			}

			noDeleted++
			if noDeleted == pending {
				break
			}
		}
	}

	return nil
}
