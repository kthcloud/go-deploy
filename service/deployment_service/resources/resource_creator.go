package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"log"
)

type SsCreatorType[T service.SsResource] struct {
	id     *string
	name   *string
	public T

	dbKey  string
	dbFunc func(string, string, interface{}) error

	createFunc func(T) (T, error)
}

func SsCreator[T service.SsResource](createFunc func(T) (T, error)) *SsCreatorType[T] {
	return &SsCreatorType[T]{
		createFunc: createFunc,
		dbFunc:     deploymentModel.New().UpdateSubsystemByID_test,
	}
}

func (rc *SsCreatorType[T]) WithID(id string) *SsCreatorType[T] {
	rc.id = &id
	return rc
}

func (rc *SsCreatorType[T]) WithDbKey(dbKey string) *SsCreatorType[T] {
	rc.dbKey = dbKey
	return rc
}

func (rc *SsCreatorType[T]) WithDbFunc(dbFunc func(string, string, interface{}) error) *SsCreatorType[T] {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsCreatorType[T]) WithPublic(public T) *SsCreatorType[T] {
	rc.public = public
	return rc
}

func (rc *SsCreatorType[T]) Exec() error {
	if rc.public == nil {
		log.Println("no public resource provided for subsystem creation. assuming it failed to create")
		return nil
	}

	if service.NotCreated(rc.public) {
		var resource T
		if rc.createFunc == nil {
			resource = rc.public
		} else {
			var err error
			resource, err = rc.createFunc(rc.public)
			if err != nil {
				return err
			}
		}

		if resource == nil {
			log.Println("no resource returned after creation. assuming it failed to create or was skipped")
		} else if rc.dbKey == "" {
			log.Println("no db key provided for subsystem creation. did you forget to call WithDbKey?")
		} else {
			if rc.id != nil {
				return rc.dbFunc(*rc.id, rc.dbKey, resource)
			} else {
				log.Println("no id or name provided for subsystem creation. did you forget to call WithID?")
			}
		}
	}

	return nil
}
