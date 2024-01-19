package gpu

import (
	"encoding/base64"
	"go-deploy/models/dto/body"
)

// ToDTO converts a GPU to a body.GpuRead DTO.
func (gpu *GPU) ToDTO(addUserInfo bool) body.GpuRead {
	id := base64.StdEncoding.EncodeToString([]byte(gpu.ID))

	var lease *body.GpuLease

	if gpu.Lease.VmID != "" {
		lease = &body.GpuLease{
			End:     gpu.Lease.End,
			Expired: gpu.Lease.IsExpired(),
		}

		if addUserInfo {
			lease.User = &gpu.Lease.UserID
			lease.VmID = &gpu.Lease.VmID
		}
	}

	return body.GpuRead{
		ID:   id,
		Name: gpu.Data.Name,
		Zone: gpu.Zone,

		Lease: lease,
	}
}
