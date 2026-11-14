package model

import (
	"time"

	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/utils"
)

//go:generate go tool goverter gen ./

// goverter:converter
// goverter:output:file ./generated.go
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend NonZeroTimeToPtr
// goverter:extend TimeIdentity
// goverter:extend AnyIdentity
type Converter interface {
	WorkerStatusToDTO(ws *WorkerStatus) body.WorkerStatusRead
	QuotasToDTO(q *Quotas) body.Quota
	HostToDTO(host *Host) body.HostRead
	HostToVerboseDTO(host *Host) body.HostVerboseRead
	NotificationToDTO(n *Notification) body.NotificationRead
	ApiKeyToCreatedDTO(apiKey *ApiKey) body.ApiKeyCreated
	ApiKeyToDTO(apiKey ApiKey) body.ApiKey
	ApiKeysToDTO(apiKeys []ApiKey) []body.ApiKey
	PublicKeyToDTO(key PublicKey) body.PublicKey
	PublicKeysToDTO(keys []PublicKey) []body.PublicKey
	PublicKeyFromDTO(key body.PublicKey) PublicKey
	PublicKeysFromDTO(keys []body.PublicKey) []PublicKey
	UserDataToDTO(data UserData) body.UserData
	UserDataSliceToDTO(data []UserData) []body.UserData
	// goverter:map GpuGroupID GpuGroupName
	GpuLeaseCreateFromDTO(dto *body.GpuLeaseCreate) GpuLeaseCreateParams
	// goverter:ignore ActivatedAt
	GpuLeaseUpdateFromDTO(dto *body.GpuLeaseUpdate) *GpuLeaseUpdateParams
	ResourceMigrationUpdateFromDTO(dto *body.ResourceMigrationUpdate) *ResourceMigrationUpdateParams
	TeamJoinFromDTO(dto *body.TeamJoin) TeamJoinParams
	JobUpdateFromDTO(dto *body.JobUpdate) JobUpdateParams
	VmActionFromDTO(dto *body.VmActionCreate) VmActionParams
	VmSnapshotToDTO(sc *SnapshotV2) body.VmSnapshotRead
	PortCreateFromDTO(port *body.PortCreate) PortCreateParams
}

var dtoConverter Converter

// NonZeroTimeToPtr converts a time.Time to *time.Time, returning nil for zero values.
func NonZeroTimeToPtr(t time.Time) *time.Time {
	return utils.NonZeroOrNil(t)
}

// TimeIdentity preserves time.Time values instead of copying only exported fields.
func TimeIdentity(t time.Time) time.Time {
	return t
}

// AnyIdentity returns the value unchanged for interface{}/any fields.
func AnyIdentity(v any) any {
	return v
}
