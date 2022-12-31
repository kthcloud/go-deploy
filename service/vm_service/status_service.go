package vm_service

import (
	"go-deploy/pkg/status_codes"
)

func GetStatusByID(userID, vmID string) (int, string, error) {
	vm, err := GetByID(userID, vmID)
	if err != nil {
		return -1, "Unknown", err
	}

	if vm == nil {
		return status_codes.ResourceNotFound, status_codes.GetMsg(status_codes.ResourceNotFound), nil
	}

	if vm.BeingDeleted {
		return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
	}

	if vm.BeingCreated {
		return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
	}

	// TODO: fetch status from cloudstack

	return status_codes.ResourceRunning, status_codes.GetMsg(status_codes.ResourceRunning), nil
}
