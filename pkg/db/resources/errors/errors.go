package errors

import "fmt"

var (
	// NonUniqueFieldErr is returned when a unique index in the database is violated.
	NonUniqueFieldErr = fmt.Errorf("non unique field")
)
