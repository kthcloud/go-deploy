package resources

import (
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems"
)

// SsUpdaterType is a type that can be used to update a single subsystem model
// It contains the public model that should be updated, and the functions that should be used to update it
type SsUpdaterType[T subsystems.SsResource] struct {
	public T

	dbFunc     func(interface{}) error
	updateFunc func(T) (T, error)
}

// SsUpdater create a new updater that can be used for a single model
func SsUpdater[T subsystems.SsResource](updateFunc func(T) (T, error)) *SsUpdaterType[T] {
	return &SsUpdaterType[T]{
		updateFunc: updateFunc,
	}
}

// WithDbFunc sets the db func for the updater
func (rc *SsUpdaterType[T]) WithDbFunc(dbFunc func(interface{}) error) *SsUpdaterType[T] {
	rc.dbFunc = dbFunc
	return rc
}

// WithPublic sets the desired public model for the updater
func (rc *SsUpdaterType[T]) WithPublic(public T) *SsUpdaterType[T] {
	rc.public = public
	return rc
}

// Exec executes the updater
func (rc *SsUpdaterType[T]) Exec() error {
	if subsystems.Nil(rc.public) {
		log.Println("No public model provided for subsystem update. did you forget to call WithPublic?")
		return nil
	}

	if subsystems.Created(rc.public) {
		var resource T
		if rc.updateFunc == nil {
			log.Println("No update function provided for subsystem update. did you forget to specify it in the constructor?")
			resource = rc.public
		} else {
			var err error
			resource, err = rc.updateFunc(rc.public)
			if err != nil {
				return err
			}
		}

		if subsystems.Nil(resource) {
			log.Println("No model returned after update. Assuming it was deleted")
			return nil
		}

		if rc.dbFunc == nil {
			log.Println("No db func provided for subsystem update. did you forget to call WithDbFunc?")
			return nil
		}

		return rc.dbFunc(resource)
	}

	return nil
}
