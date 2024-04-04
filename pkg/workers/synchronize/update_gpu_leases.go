package synchronize

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/gpu_group_repo"
	"go-deploy/pkg/db/resources/gpu_lease_repo"
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

	// 1. Check for expired leases
	err = checkForExpiredLeases(gpuLeases)
	if err != nil {
		return err
	}

	// 2. Check for leases that has a queue position of 0 without being assigned (They should be assigned)
	err = checkForUnassignedLeases(gpuLeases, gpuGroups)
	if err != nil {
		return err
	}

	// 3. Check for leases that are assigned 24 hours ago and are not yet activated
	err = checkForUnactivatedLeases(gpuLeases)
	if err != nil {
		return err
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
