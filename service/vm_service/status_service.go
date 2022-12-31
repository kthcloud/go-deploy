package vm_service

import (
	"go-deploy/pkg/status_codes"
	"go-deploy/service/vm_service/internal_service"
)

func GetStatusByID(userID, vmID string) (int, string, error) {
	vm, err := GetByID(userID, vmID)
	if err != nil {
		return -1, "Unknown", err
	}

	csStatusCode, csStatusMsg, err := internal_service.GetStatusCS(vm.Name)
	if err != nil || csStatusCode == status_codes.ResourceUnknown {
		if vm.BeingDeleted {
			return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
		}

		if vm.BeingCreated {
			return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
		}

		return status_codes.ResourceUnknown, status_codes.GetMsg(status_codes.ResourceUnknown), nil
	}

	return csStatusCode, csStatusMsg, nil
}
