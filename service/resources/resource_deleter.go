package resources

import (
	"go-deploy/pkg/log"
)

// SsDeleterType is a type that can be used to delete a single subsystem model
// It contains the model ID that should be deleted, and the functions that should be used to delete it
type SsDeleterType[IdType any] struct {
	resourceID *IdType

	dbFunc     func(interface{}) error
	deleteFunc func(IdType) error
}

// SsDeleter create a new deleter that can be used for a single model
func SsDeleter[IdType any](deleteFunc func(IdType) error) *SsDeleterType[IdType] {
	return &SsDeleterType[IdType]{
		deleteFunc: deleteFunc,
	}
}

// WithResourceID sets the model ID for the deleter.
// The model ID refers to the subsystem model's ID, not the deployment ID, VM ID, or SM ID
func (rc *SsDeleterType[IdType]) WithResourceID(resourceID IdType) *SsDeleterType[IdType] {
	rc.resourceID = &resourceID
	return rc
}

// WithDbFunc sets the db func for the deleter
func (rc *SsDeleterType[IdType]) WithDbFunc(dbFunc func(interface{}) error) *SsDeleterType[IdType] {
	rc.dbFunc = dbFunc
	return rc
}

// Exec executes the deleter
func (rc *SsDeleterType[IdType]) Exec() error {
	if rc.resourceID == nil {
		log.Println("no model id provided for subsystem deletion. did you forget to call WithResourceID?")
	} else if rc.deleteFunc == nil {
		log.Println("no delete function provided for subsystem deletion. did you forget to specify it in the constructor?")
	} else {
		err := rc.deleteFunc(*rc.resourceID)
		if err != nil {
			return err
		}
	}

	if rc.dbFunc == nil {
		log.Println("no db func provided for subsystem deletion. did you forget to call WithDbFunc?")
		return nil
	}

	return rc.dbFunc(nil)
}
