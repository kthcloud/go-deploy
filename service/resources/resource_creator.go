package resources

import (
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems"
)

// SsCreatorType is a type that can be used to create a single subsystem model
// It contains the public model that should be created, and the functions that should be used to create it
type SsCreatorType[T subsystems.SsResource] struct {
	name   *string
	public T

	dbFunc     func(interface{}) error
	createFunc func(T) (T, error)
}

// SsCreator create a new creator that can be used for a single model
func SsCreator[T subsystems.SsResource](createFunc func(T) (T, error)) *SsCreatorType[T] {
	return &SsCreatorType[T]{
		createFunc: createFunc,
	}
}

// WithDbFunc sets the db func for the creator
func (rc *SsCreatorType[T]) WithDbFunc(dbFunc func(interface{}) error) *SsCreatorType[T] {
	rc.dbFunc = dbFunc
	return rc
}

// WithPublic sets the public model for the creator
func (rc *SsCreatorType[T]) WithPublic(public T) *SsCreatorType[T] {
	rc.public = public
	return rc
}

// Exec executes the creator
func (rc *SsCreatorType[T]) Exec() error {
	if subsystems.Nil(rc.public) {
		log.Println("No public model provided for subsystem creation. Assuming it failed to create")
		return nil
	}

	if subsystems.NotCreated(rc.public) {
		var resource T
		if rc.createFunc == nil {
			log.Println("No create function provided for subsystem creation. did you forget to specify it in the constructor?")
			return nil
		} else {
			var err error
			resource, err = rc.createFunc(rc.public)
			if err != nil {
				return err
			}
		}

		if subsystems.Nil(resource) {
			log.Println("No model returned after creation. Assuming it failed to create or was skipped")
			return nil
		}

		if rc.dbFunc == nil {
			log.Println("No db func provided for subsystem creation. did you forget to call WithDbFunc?")
			return nil
		}

		subsystems.CopyValue(resource, rc.public)
		return rc.dbFunc(resource)
	}

	return nil
}
