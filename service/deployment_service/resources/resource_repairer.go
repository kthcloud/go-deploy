package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"log"
)

type SsRepairerType[idType any, T service.SsResource] struct {
	id         *string
	name       *string
	resourceID *idType
	dbKey      string
	genPublic  func() T

	fetchFunc  func(idType) (T, error)
	createFunc func(T) (T, error)
	updateFunc func(T) (T, error)
	deleteFunc func(idType) error
}

func SsRepairer[idType any, T service.SsResource](
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

func (rc *SsRepairerType[idType, T]) WithID(id string) *SsRepairerType[idType, T] {
	rc.id = &id
	return rc
}

func (rc *SsRepairerType[idType, T]) WithName(name string) *SsRepairerType[idType, T] {
	rc.name = &name
	return rc
}

func (rc *SsRepairerType[idType, T]) WithResourceID(resourceID idType) *SsRepairerType[idType, T] {
	rc.resourceID = &resourceID
	return rc
}

func (rc *SsRepairerType[idType, T]) WithDbKey(dbKey string) *SsRepairerType[idType, T] {
	rc.dbKey = dbKey
	return rc
}

func (rc *SsRepairerType[idType, T]) WithGenPublicFunc(genPublic func() T) *SsRepairerType[idType, T] {
	rc.genPublic = genPublic
	return rc
}

func (rc *SsRepairerType[idType, T]) Exec() error {
	dbResource, err := rc.fetchFunc(*rc.resourceID)
	if err != nil {
		return err
	}

	if service.NotCreated(dbResource) {
		return rc.createResourceInstead()
	}

	return service.UpdateIfDiff[T](
		dbResource,
		func() (T, error) {
			return rc.fetchFunc(*rc.resourceID)
		},
		rc.updateFunc,
		func(dbResource T) error {
			err := rc.deleteFunc(*rc.resourceID)
			if err != nil {
				return err
			}

			_, err = rc.createFunc(dbResource)
			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (rc *SsRepairerType[idType, T]) createResourceInstead() error {
	if rc.genPublic == nil {
		log.Println("no gen public function for subsystem repair. did you forget to call WithGenPublicFunc?")
		return nil
	}

	if rc.createFunc == nil {
		log.Println("no create function provided for subsystem repair. did you forget to specify it in the constructor?")
		return nil
	}

	public := rc.genPublic()
	if public == nil {
		log.Println("no public supplied for subsystem repair. assuming it failed to create")
		return nil
	}

	resource, err := rc.createFunc(public)
	if err != nil {
		return err
	}

	if resource == nil {
		log.Println("no resource returned after creation. assuming it failed to create or was skipped")
		return nil
	}

	if rc.dbKey == "" {
		log.Println("no db key provided for subsystem creation. did you forget to call WithDbKey?")
		return nil
	}

	if rc.id != nil {
		return deploymentModel.New().UpdateSubsystemByID_test(*rc.id, rc.dbKey, resource)
	}

	if rc.name != nil {
		return deploymentModel.New().UpdateSubsystemByName_test(*rc.name, rc.dbKey, resource)
	}

	log.Println("no id or name provided for subsystem update. did you forget to call WithID or WithName?")
	return nil
}
