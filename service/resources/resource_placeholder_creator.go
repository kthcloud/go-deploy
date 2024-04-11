package resources

import (
	"go-deploy/pkg/log"
)

// SsPlaceholderCreatorType is a type that can be used to create a single subsystem placeholder model.
// It contains the functions that should be used to create it.
// A placeholder subsystem creator is essentially a normal subsystem creator without any public model
type SsPlaceholderCreatorType struct {
	dbFunc func(interface{}) error
}

// SsPlaceholderCreator create a new creator that can be used for a single model
func SsPlaceholderCreator() *SsPlaceholderCreatorType {
	return &SsPlaceholderCreatorType{}
}

// WithDbFunc sets the db func for the creator
func (rc *SsPlaceholderCreatorType) WithDbFunc(dbFunc func(interface{}) error) *SsPlaceholderCreatorType {
	rc.dbFunc = dbFunc
	return rc
}

// Exec executes the placeholder creator
func (rc *SsPlaceholderCreatorType) Exec() error {
	if rc.dbFunc == nil {
		log.Println("No db key provided for subsystem placeholder creation. did you forget to call WithDbKey?")
		return nil
	}

	return rc.dbFunc(true)
}
