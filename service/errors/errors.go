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
	// ErrAuthInfoNotAvailable is returned when the auth info is not available
	ErrAuthInfoNotAvailable = fmt.Errorf("auth info not available")

	// ErrBadDiscoveryToken is returned when the discovery token is invalid.
	ErrBadDiscoveryToken = fmt.Errorf("bad discovery token")

	// ErrDeploymentNotFound is returned when the deployment is not found.
	// This is most likely caused by a race-condition between a some model call and a deletion call.
	ErrDeploymentNotFound = fmt.Errorf("deployment not found")

	// ErrDeploymentHasNoCiConfig is returned when the deployment does not have a CI config.
	ErrDeploymentHasNoCiConfig = fmt.Errorf("deployment does not have a CI config")

	// ErrMainAppNotFound is returned when the main app is not found.
	// This could be caused by stale data in the database.
	ErrMainAppNotFound = fmt.Errorf("main app not found")

	// ErrIngressHostInUse is returned when the ingress host is already in use.
	ErrIngressHostInUse = fmt.Errorf("ingress host already in use")

	// ErrSmNotFound is returned when the storage manager is not found.
	ErrSmNotFound = fmt.Errorf("storage manager not found")

	// ErrSmAlreadyExists is returned when the storage manager already exists for user.
	ErrSmAlreadyExists = fmt.Errorf("storage manager already exists for user")

	// ErrVmNotFound is returned when the vm is not found.
	// This is most likely caused by a race-condition between a some model call and a deletion call.
	ErrVmNotFound = fmt.Errorf("vm not found")

	// ErrGpuNotFound is returned when the gpu is not found.
	ErrGpuNotFound = fmt.Errorf("gpu not found")

	// ErrGpuLeaseAlreadyExists is returned when a GPU lease already exists for a user.
	ErrGpuLeaseAlreadyExists = fmt.Errorf("gpu lease already exists")

	// ErrGpuLeaseNotActive is returned when the GPU lease is not active.
	ErrGpuLeaseNotActive = fmt.Errorf("gpu lease not active")

	// ErrGpuLeaseNotAssigned is returned when the GPU lease is not assigned.
	ErrGpuLeaseNotAssigned = fmt.Errorf("gpu lease not assigned")

	// ErrGpuLeaseNotFound is returned when the GPU lease is not found.
	ErrGpuLeaseNotFound = fmt.Errorf("gpu lease not found")

	// ErrVmAlreadyAttached is returned when a VM is already attached to a GPU lease.
	ErrVmAlreadyAttached = fmt.Errorf("vm already attached")

	// ErrGpuGroupNotFound is returned when the GPU group is not found.
	ErrGpuGroupNotFound = fmt.Errorf("gpu group not found")

	// ErrNoPortsAvailable is returned when there are no ports available.
	ErrNoPortsAvailable = fmt.Errorf("no ports available")

	// ErrZoneNotFound is returned when the zone is not found.
	// This could be caused by stale data in the database.
	ErrZoneNotFound = fmt.Errorf("zone not found")

	// ErrHostNotFound is returned when the host is not found.
	// This could be caused by stale data in the database or inconsistent state of the system/cluster.
	ErrHostNotFound = fmt.Errorf("host not found")

	// ErrSnapshotNotFound is returned when the snapshot is not found.
	ErrSnapshotNotFound = fmt.Errorf("snapshot not found")

	// ErrNonUniqueField is returned when a field is not unique, such as the name of a deployment.
	ErrNonUniqueField = fmt.Errorf("non unique field")

	// ErrTeamNameTaken is returned when the team name is already taken by another team.
	// Every team name should be unique.
	ErrTeamNameTaken = fmt.Errorf("team name taken")

	// ErrTeamNotFound is returned when the team is not found.
	ErrTeamNotFound = fmt.Errorf("team not found")

	// ErrUserNotFound is returned when the user is not found.
	ErrUserNotFound = fmt.Errorf("user not found")

	// ErrApiKeyNameTaken is returned when the API key name is already taken by another API key.
	// Every API key name should be unique.
	ErrApiKeyNameTaken = fmt.Errorf("api key name taken")

	// ErrBadInviteCode is returned when the invite code is invalid.
	ErrBadInviteCode = fmt.Errorf("bad invite code")

	// ErrNotInvited is returned when the user tries to join a team, but is not invited.
	ErrNotInvited = fmt.Errorf("not invited")

	// ErrForbidden is returned when the user is not allowed to perform an action.
	ErrForbidden = fmt.Errorf("forbidden")

	// ErrJobNotFound is returned when the job is not found.
	ErrJobNotFound = fmt.Errorf("job not found")

	// ErrNotificationNotFound is returned when the notification is not found.
	ErrNotificationNotFound = fmt.Errorf("notification not found")

	// ErrResourceMigrationNotFound is returned when the resource migration is not found.
	ErrResourceMigrationNotFound = fmt.Errorf("resource migration not found")

	// ErrAlreadyAccepted is returned when the resource migration is already accepted.
	ErrAlreadyAccepted = fmt.Errorf("resource migration already accepted")

	// ErrAlreadyMigrated is returned when the resource is already migrated.
	ErrAlreadyMigrated = fmt.Errorf("resource already migrated")

	// ErrResourceMigrationAlreadyExists is returned when the resource migration already exists.
	ErrResourceMigrationAlreadyExists = fmt.Errorf("resource migration already exists")

	// ErrBadResourceMigrationType is returned when the resource migration type is invalid.
	// This could be caused by providing another type than what is expected by the function.
	ErrBadResourceMigrationType = fmt.Errorf("bad resource migration type")

	// ErrBadResourceMigrationResourceType is returned when the resource migration resource type is invalid.
	// This could be caused by providing another type than what is expected by the function.
	ErrBadResourceMigrationResourceType = fmt.Errorf("bad resource migration resource type")

	// ErrBadResourceMigrationParamsErr is returned when the resource migration params are invalid.
	// This could be caused by providing another type than what is expected by the function.
	ErrBadResourceMigrationParams = fmt.Errorf("bad resource migration params")

	// ErrBadResourceMigrationStatus is returned when the resource migration status is invalid.
	// This could be caused by providing another type than what is expected by the function.
	ErrBadResourceMigrationStatus = fmt.Errorf("bad resource migration status")

	// ErrResourceNotFound is returned when the resource is not found.
	// This is used when the type of resource is unknown, such as for team resources or migration resources.
	ErrResourceNotFound = fmt.Errorf("resource not found")

	// ErrBadMigrationCode is returned when the migration code is invalid.
	ErrBadMigrationCode = fmt.Errorf("bad migration code")
)
