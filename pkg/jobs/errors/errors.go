package errors

import "fmt"

// MakeTerminatedError makes a terminated error.
func MakeTerminatedError(err error) error {
	return fmt.Errorf("terminated: %w", err)
}

// MakeFailedError makes a failed error.
func MakeFailedError(err error) error {
	return fmt.Errorf("failed: %w", err)
}
