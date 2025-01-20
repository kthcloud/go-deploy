package resources

import (
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems"
	"github.com/kthcloud/go-deploy/service/utils"
)

// SsRepairerType is a type that can be used to repair a single subsystem model
// It contains the public model that should be created,
// and the functions that should be used to create, update and delete it.
//
// A SsRepairerType works by inspecting the diff between the database model and the live model,
// then if there is any diff, it will try to update the live model.
// Then it will compare again, and if there is still a diff, it will delete the live model and create a new one.
//
// It make use of service.UpdateIfDiff to achieve this.
type SsRepairerType[idType any, T subsystems.SsResource] struct {
	resourceID *idType

	genPublic     T
	genPublicFunc func() T

	dbFunc     func(interface{}) error
	fetchFunc  func(idType) (T, error)
	createFunc func(T) (T, error)
	updateFunc func(T) (T, error)
	deleteFunc func(idType) error
}

// SsRepairer create a new repairer that can be used for a single model
func SsRepairer[idType any, T subsystems.SsResource](
	fetchFunc func(idType) (T, error),
	createFunc func(T) (T, error),
	updateFunc func(T) (T, error),
	deleteFunc func(idType) error,
) *SsRepairerType[idType, T] {
	return &SsRepairerType[idType, T]{
		fetchFunc:  fetchFunc,
		createFunc: createFunc,
		updateFunc: updateFunc,
		deleteFunc: deleteFunc,
	}
}

// WithResourceID sets the model ID for the repairer.
// The model ID refers to the subsystem model's ID, not the deployment ID, VM ID, or SM ID
func (rc *SsRepairerType[idType, T]) WithResourceID(resourceID idType) *SsRepairerType[idType, T] {
	rc.resourceID = &resourceID
	return rc
}

// WithDbFunc sets the db func for the repairer
func (rc *SsRepairerType[idType, T]) WithDbFunc(dbFunc func(interface{}) error) *SsRepairerType[idType, T] {
	rc.dbFunc = dbFunc
	return rc
}

// WithGenPublic sets the desired public model for the repairer
// This is only used if there is a diff between the database model and the live model
func (rc *SsRepairerType[idType, T]) WithGenPublic(genPublic T) *SsRepairerType[idType, T] {
	rc.genPublic = genPublic
	return rc
}

// WithGenPublicFunc sets the desired public model for the repairer
// This is only called if there is a diff between the database model and the live model
func (rc *SsRepairerType[idType, T]) WithGenPublicFunc(genPublicFunc func() T) *SsRepairerType[idType, T] {
	rc.genPublicFunc = genPublicFunc
	return rc
}

// Exec executes the repairer
func (rc *SsRepairerType[idType, T]) Exec() error {
	var dbResource T
	if rc.genPublicFunc != nil {
		dbResource = rc.genPublicFunc()
	} else if subsystems.NotNil(rc.genPublic) {
		dbResource = rc.genPublic
	} else {
		log.Println("No genPublic or genPublicFunc provided for subsystem repair. did you forget to call WithGenPublic or WithGenPublicFunc?")
		return nil
	}

	if subsystems.NotCreated(dbResource) {
		return rc.createResourceInstead()
	}

	return utils.UpdateIfDiff[T](
		dbResource,
		func() (T, error) {
			return rc.fetchFunc(*rc.resourceID)
		},
		func(dbResource T) (T, error) {
			var empty T

			updated, err := rc.updateFunc(dbResource)
			if err != nil {
				return empty, err
			}

			if rc.dbFunc == nil {
				log.Println("No db func provided for subsystem repair. did you forget to call WithDbFunc?")
				return updated, nil
			}

			err = rc.dbFunc(updated)
			if err != nil {
				return empty, err
			}

			return updated, nil
		},
		func(dbResource T) error {
			err := rc.deleteFunc(*rc.resourceID)
			if err != nil {
				return err
			}

			created, err := rc.createFunc(dbResource)
			if err != nil {
				return err
			}

			if rc.dbFunc == nil {
				log.Println("No db func provided for subsystem repair. did you forget to call WithDbFunc?")
				return nil
			}

			err = rc.dbFunc(created)
			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (rc *SsRepairerType[idType, T]) createResourceInstead() error {

	if rc.createFunc == nil {
		log.Println("No create function provided for subsystem repair. did you forget to specify it in the constructor?")
		return nil
	}

	var public T
	if rc.genPublicFunc != nil {
		public = rc.genPublicFunc()
	} else if subsystems.NotNil(rc.genPublic) {
		public = rc.genPublic
	} else {
		log.Println("No genPublic or genPublicFunc provided for subsystem repair. did you forget to call WithGenPublic or WithGenPublicFunc?")
		return nil
	}

	if subsystems.Nil(public) {
		log.Println("No public supplied for subsystem repair. Assuming it failed to create or was skipped")
		return nil
	}

	resource, err := rc.createFunc(public)
	if err != nil {
		return err
	}

	if subsystems.Nil(resource) {
		log.Println("No model returned after creation. Assuming it failed to create or was skipped")
		return nil
	}

	if rc.dbFunc == nil {
		log.Println("No db func provided for subsystem repair. did you forget to call WithDbFunc?")
		return nil
	}

	return rc.dbFunc(resource)
}
