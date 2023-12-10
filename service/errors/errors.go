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

	// StorageManagerNotFoundErr is returned when the storage manager is not found.
	StorageManagerNotFoundErr = fmt.Errorf("storage manager not found")

	// StorageManagerAlreadyExistsErr is returned when the storage manager already exists for user.
	StorageManagerAlreadyExistsErr = fmt.Errorf("storage manager already exists for user")

	// ZoneNotFoundErr is returned when the zone is not found.
	// This could be caused by stale data in the database.
	ZoneNotFoundErr = fmt.Errorf("zone not found")

	// CustomDomainInUseErr is returned when the custom domain is already in use.
	// This is most likely a race condition, where two resources request the same custom domain.
	CustomDomainInUseErr = fmt.Errorf("custom domain is already in use")

	// NonUniqueFieldErr is returned when a field is not unique, such as the name of a deployment.
	NonUniqueFieldErr = fmt.Errorf("non unique field")

	// InvalidTransferCodeErr is returned when the transfer code is invalid.
	// This could be caused by the transfer being canceled, or the code being incorrect.
	InvalidTransferCodeErr = fmt.Errorf("invalid transfer code")
)
