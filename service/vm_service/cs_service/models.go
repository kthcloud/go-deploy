package cs_service

import csModels "go-deploy/pkg/subsystems/cs/models"

const (
	CsDetachGpuAfterStateRestore = "restore"
	CsDetachGpuAfterStateOff     = "off"
	CsDetachGpuAfterStateOn      = "on"
)

type CsCreated struct {
	VM *csModels.VmPublic
}
