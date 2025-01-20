package errors

import "fmt"

var (
	// ErrNonUniqueField is returned when a unique index in the database is violated.
	ErrNonUniqueField = fmt.Errorf("non unique field")
)
