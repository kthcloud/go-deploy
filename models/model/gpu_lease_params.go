package model

import "time"

type GpuLeaseUpdateParams struct {
	ActivatedAt *time.Time
	VmID        *string
}
