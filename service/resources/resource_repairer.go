package resources

import (
	"go-deploy/service"
	"log"
)

type SsRepairerType[idType any, T service.SsResource] struct {
	resourceID *idType

	genPublic     T
	genPublicFunc func() T

	dbFunc     func(interface{}) error
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

func (rc *SsRepairerType[idType, T]) WithResourceID(resourceID idType) *SsRepairerType[idType, T] {
	rc.resourceID = &resourceID
	return rc
}

func (rc *SsRepairerType[idType, T]) WithDbFunc(dbFunc func(interface{}) error) *SsRepairerType[idType, T] {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsRepairerType[idType, T]) WithGenPublic(genPublic T) *SsRepairerType[idType, T] {
	rc.genPublic = genPublic
	return rc
}

func (rc *SsRepairerType[idType, T]) WithGenPublicFunc(genPublicFunc func() T) *SsRepairerType[idType, T] {
	rc.genPublicFunc = genPublicFunc
	return rc
}

func (rc *SsRepairerType[idType, T]) Exec() error {
	var dbResource T
	if rc.genPublicFunc != nil {
		dbResource = rc.genPublicFunc()
	} else if service.NotNil(rc.genPublic) {
		dbResource = rc.genPublic
	} else {
		log.Println("no genPublic or genPublicFunc provided for subsystem repair. did you forget to call WithGenPublic or WithGenPublicFunc?")
		return nil
	}

	if service.NotCreated(dbResource) {
		return rc.createResourceInstead()
	}

	return service.UpdateIfDiff[T](
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
				log.Println("no db func provided for subsystem repair. did you forget to call WithDbFunc?")
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
				log.Println("no db func provided for subsystem repair. did you forget to call WithDbFunc?")
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
		log.Println("no create function provided for subsystem repair. did you forget to specify it in the constructor?")
		return nil
	}

	var public T
	if rc.genPublicFunc != nil {
		public = rc.genPublicFunc()
	} else if service.NotNil(rc.genPublic) {
		public = rc.genPublic
	} else {
		log.Println("no genPublic or genPublicFunc provided for subsystem repair. did you forget to call WithGenPublic or WithGenPublicFunc?")
		return nil
	}

	if service.Nil(public) {
		log.Println("no public supplied for subsystem repair. assuming it failed to create or was skipped")
		return nil
	}

	resource, err := rc.createFunc(public)
	if err != nil {
		return err
	}

	if service.Nil(resource) {
		log.Println("no resource returned after creation. assuming it failed to create or was skipped")
		return nil
	}

	if rc.dbFunc == nil {
		log.Println("no db func provided for subsystem repair. did you forget to call WithDbFunc?")
		return nil
	}

	return rc.dbFunc(resource)
}
