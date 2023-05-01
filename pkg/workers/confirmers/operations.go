package confirmers

import (
	"fmt"
	"go-deploy/models/vm"
	"go-deploy/service/vm_service"
)

func ReturnGPU(gpu *vm.GPU) error {
	if gpu.Lease.VmID == "" {
		return nil
	}

	err := vm_service.DetachGpuSync(gpu.Lease.VmID, gpu.Lease.UserID)
	if err != nil {
		return fmt.Errorf("failed to return gpu %s. details: %s", gpu.ID, err)
	}

	return nil
}
