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

var (
	// DeploymentNotFoundErr is returned when the deployment is not found.
	// This is most likely caused by a race-condition between a some resource call and a deletion call.
	DeploymentNotFoundErr = fmt.Errorf("deployment not found")

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
	// This is most likely caused by a race-condition between a some resource call and a deletion call.
	VmNotFoundErr = fmt.Errorf("vm not found")

	// VmNotCreatedErr is returned when the vm is not created.
	VmNotCreatedErr = fmt.Errorf("vm not created")

	// GpuNotFoundErr is returned when the gpu is not found.
	GpuNotFoundErr = fmt.Errorf("gpu not found")

	// GpuAlreadyAttachedErr is returned when the gpu is already attached to another VM.
	GpuAlreadyAttachedErr = fmt.Errorf("gpu already attached")

	// VmTooLargeErr is returned when the VM is too large to be started on a specific host.
	// Something that is required when using GPUs.
	VmTooLargeErr = fmt.Errorf("vm too large")

	// PortInUseErr is returned when the port is already in use.
	PortInUseErr = fmt.Errorf("port in use")

	// HostNotAvailableErr is returned when the host is not available.
	// This is usually caused by the host being in maintenance or disabled mode.
	HostNotAvailableErr = fmt.Errorf("host not available")

	// ZoneNotFoundErr is returned when the zone is not found.
	// This could be caused by stale data in the database.
	ZoneNotFoundErr = fmt.Errorf("zone not found")

	// NonUniqueFieldErr is returned when a field is not unique, such as the name of a deployment.
	NonUniqueFieldErr = fmt.Errorf("non unique field")

	// InvalidTransferCodeErr is returned when the transfer code is invalid.
	// This could be caused by the transfer being canceled, or the code being incorrect.
	InvalidTransferCodeErr = fmt.Errorf("invalid transfer code")

	// TeamNameTakenErr is returned when the team name is already taken by another team.
	// Every team name should be unique.
	TeamNameTakenErr = fmt.Errorf("team name taken")

	// TeamNotFoundErr is returned when the team is not found.
	TeamNotFoundErr = fmt.Errorf("team not found")

	// UserNotFoundErr is returned when the user is not found.
	UserNotFoundErr = fmt.Errorf("user not found")

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
)
