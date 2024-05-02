package errors

import "fmt"

// QuotaExceededError is returned when the quota is exceeded.
// For instance, if a user creates too many deployments.
type QuotaExceededError struct {
	reason string
}

// Error returns the reason for the quota exceeded error.
func (e QuotaExceededError) Error() string {
	return e.reason
}

// NewQuotaExceededError creates a new QuotaExceededError.
func NewQuotaExceededError(reason string) QuotaExceededError {
	return QuotaExceededError{reason: reason}
}

// FailedToStartActivityError is returned when an activity fails to start.
// For instance, if a user tries to update a deployment, that is being deleted.
type FailedToStartActivityError struct {
	reason string
}

// Error returns the reason for the failed to start activity error.
func (e FailedToStartActivityError) Error() string {
	return e.reason
}

// NewFailedToStartActivityError creates a new FailedToStartActivityError.
func NewFailedToStartActivityError(reason string) FailedToStartActivityError {
	return FailedToStartActivityError{reason: reason}
}

// PortInUseErr is returned when the port is already in use.
type PortInUseErr struct {
	Port int
}

// Error returns the reason for the port in use error.
func (e PortInUseErr) Error() string {
	return fmt.Sprintf("port %d is already in use", e.Port)
}

// NewPortInUseError creates a new PortInUseErr.
func NewPortInUseError(port int) PortInUseErr {
	return PortInUseErr{Port: port}
}

// ZoneCapabilityMissingErr is returned when the zone capability is missing.
type ZoneCapabilityMissingErr struct {
	Zone       string
	Capability string
}

// Error returns the reason for the zone capability missing error.
func (e ZoneCapabilityMissingErr) Error() string {
	return fmt.Sprintf("zone %s is missing capability %s", e.Zone, e.Capability)
}

// NewZoneCapabilityMissingError creates a new ZoneCapabilityMissingErr.
func NewZoneCapabilityMissingError(zone, capability string) ZoneCapabilityMissingErr {
	return ZoneCapabilityMissingErr{Zone: zone, Capability: capability}
}

var (
	// AuthInfoNotAvailableErr is returned when the auth info is not available
	AuthInfoNotAvailableErr = fmt.Errorf("auth info not available")

	// DeploymentNotFoundErr is returned when the deployment is not found.
	// This is most likely caused by a race-condition between a some model call and a deletion call.
	DeploymentNotFoundErr = fmt.Errorf("deployment not found")

	// DeploymentHasNotCiConfigErr is returned when the deployment does not have a CI config.
	DeploymentHasNotCiConfigErr = fmt.Errorf("deployment does not have a CI config")

	// MainAppNotFoundErr is returned when the main app is not found.
	// This could be caused by stale data in the database.
	MainAppNotFoundErr = fmt.Errorf("main app not found")

	// IngressHostInUseErr is returned when the ingress host is already in use.
	IngressHostInUseErr = fmt.Errorf("ingress host already in use")

	// SmNotFoundErr is returned when the storage manager is not found.
	SmNotFoundErr = fmt.Errorf("storage manager not found")

	// SmAlreadyExistsErr is returned when the storage manager already exists for user.
	SmAlreadyExistsErr = fmt.Errorf("storage manager already exists for user")

	// VmNotFoundErr is returned when the vm is not found.
	// This is most likely caused by a race-condition between a some model call and a deletion call.
	VmNotFoundErr = fmt.Errorf("vm not found")

	// VmNotCreatedErr is returned when the vm is not created.
	VmNotCreatedErr = fmt.Errorf("vm not created")

	// GpuNotFoundErr is returned when the gpu is not found.
	GpuNotFoundErr = fmt.Errorf("gpu not found")

	// GpuAlreadyAttachedErr is returned when the GPU is already attached to another VM.
	GpuAlreadyAttachedErr = fmt.Errorf("gpu already attached")

	// GpuLeaseAlreadyExistsErr is returned when a GPU lease already exists for a user.
	GpuLeaseAlreadyExistsErr = fmt.Errorf("gpu lease already exists")

	// GpuLeaseNotActiveErr is returned when the GPU lease is not active.
	GpuLeaseNotActiveErr = fmt.Errorf("gpu lease not active")

	// GpuLeaseNotAssignedErr is returned when the GPU lease is not assigned.
	GpuLeaseNotAssignedErr = fmt.Errorf("gpu lease not assigned")

	// GpuLeaseNotFoundErr is returned when the GPU lease is not found.
	GpuLeaseNotFoundErr = fmt.Errorf("gpu lease not found")

	// GpuGroupNotFoundErr is returned when the GPU group is not found.
	GpuGroupNotFoundErr = fmt.Errorf("gpu group not found")

	// VmTooLargeErr is returned when the VM is too large to be started on a specific host.
	// Something that is required when using GPUs.
	VmTooLargeErr = fmt.Errorf("vm too large")

	// NoPortsAvailableErr is returned when there are no ports available.
	NoPortsAvailableErr = fmt.Errorf("no ports available")

	// HostNotAvailableErr is returned when the host is not available.
	// This is usually caused by the host being in maintenance or disabled mode.
	HostNotAvailableErr = fmt.Errorf("host not available")

	// ZoneNotFoundErr is returned when the zone is not found.
	// This could be caused by stale data in the database.
	ZoneNotFoundErr = fmt.Errorf("zone not found")

	// SnapshotNotFoundErr is returned when the snapshot is not found.
	SnapshotNotFoundErr = fmt.Errorf("snapshot not found")

	// SnapshotAlreadyExistsErr is returned when the snapshot already exists.
	// This is normally caused by not specifying the "overwrite" flag and a snapshot with the same name already exists.
	SnapshotAlreadyExistsErr = fmt.Errorf("already exists")

	// BadStateErr is returned when the state is bad.
	// For example, if a VM has a GPU attached when creating a snapshot or if a deployment is not ready to set up a log stream.
	BadStateErr = fmt.Errorf("bad state")

	// NonUniqueFieldErr is returned when a field is not unique, such as the name of a deployment.
	NonUniqueFieldErr = fmt.Errorf("non unique field")

	// TeamNameTakenErr is returned when the team name is already taken by another team.
	// Every team name should be unique.
	TeamNameTakenErr = fmt.Errorf("team name taken")

	// TeamNotFoundErr is returned when the team is not found.
	TeamNotFoundErr = fmt.Errorf("team not found")

	// UserNotFoundErr is returned when the user is not found.
	UserNotFoundErr = fmt.Errorf("user not found")

	// ApiKeyNameTakenErr is returned when the API key name is already taken by another API key.
	// Every API key name should be unique.
	ApiKeyNameTakenErr = fmt.Errorf("api key name taken")

	// BadInviteCodeErr is returned when the invite code is invalid.
	BadInviteCodeErr = fmt.Errorf("bad invite code")

	// NotInvitedErr is returned when the user tries to join a team, but is not invited.
	NotInvitedErr = fmt.Errorf("not invited")

	// ForbiddenErr is returned when the user is not allowed to perform an action.
	ForbiddenErr = fmt.Errorf("forbidden")

	// JobNotFoundErr is returned when the job is not found.
	JobNotFoundErr = fmt.Errorf("job not found")

	// NotificationNotFoundErr is returned when the notification is not found.
	NotificationNotFoundErr = fmt.Errorf("notification not found")

	// ResourceMigrationNotFoundErr is returned when the resource migration is not found.
	ResourceMigrationNotFoundErr = fmt.Errorf("resource migration not found")

	// AlreadyAcceptedErr is returned when the resource migration is already accepted.
	AlreadyAcceptedErr = fmt.Errorf("resource migration already accepted")

	// AlreadyMigratedErr is returned when the resource is already migrated.
	AlreadyMigratedErr = fmt.Errorf("resource already migrated")

	// ResourceMigrationAlreadyExistsErr is returned when the resource migration already exists.
	ResourceMigrationAlreadyExistsErr = fmt.Errorf("resource migration already exists")

	// BadResourceMigrationTypeErr is returned when the resource migration type is invalid.
	// This could be caused by providing another type than what is expected by the function.
	BadResourceMigrationTypeErr = fmt.Errorf("bad resource migration type")

	// BadResourceMigrationResourceTypeErr is returned when the resource migration resource type is invalid.
	// This could be caused by providing another type than what is expected by the function.
	BadResourceMigrationResourceTypeErr = fmt.Errorf("bad resource migration resource type")

	// BadResourceMigrationParamsErr is returned when the resource migration params are invalid.
	// This could be caused by providing another type than what is expected by the function.
	BadResourceMigrationParamsErr = fmt.Errorf("bad resource migration params")

	// BadResourceMigrationStatusErr is returned when the resource migration status is invalid.
	// This could be caused by providing another type than what is expected by the function.
	BadResourceMigrationStatusErr = fmt.Errorf("bad resource migration status")

	// ResourceNotFoundErr is returned when the resource is not found.
	// This is used when the type of resource is unknown, such as for team resources or migration resources.
	ResourceNotFoundErr = fmt.Errorf("resource not found")

	// BadMigrationCodeErr is returned when the migration code is invalid.
	BadMigrationCodeErr = fmt.Errorf("bad migration code")
)
