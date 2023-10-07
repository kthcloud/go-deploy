package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"log"
)

type SsCreatorType[T service.SsResource] struct {
	id         *string
	name       *string
	public     T
	dbKey      string
	createFunc func(T) (T, error)
}

func SsCreator[T service.SsResource](createFunc func(T) (T, error)) *SsCreatorType[T] {
	return &SsCreatorType[T]{
		createFunc: createFunc,
	}
}

func (rc *SsCreatorType[T]) WithID(id string) *SsCreatorType[T] {
	rc.id = &id
	return rc
}

func (rc *SsCreatorType[T]) WithName(name string) *SsCreatorType[T] {
	rc.name = &name
	return rc
}

func (rc *SsCreatorType[T]) WithDbKey(dbKey string) *SsCreatorType[T] {
	rc.dbKey = dbKey
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
				return deploymentModel.New().UpdateSubsystemByID_test(*rc.id, rc.dbKey, resource)
			} else if rc.name != nil {
				return deploymentModel.New().UpdateSubsystemByName_test(*rc.name, rc.dbKey, resource)
			} else {
				log.Println("no id or name provided for subsystem creation. did you forget to call WithID or WithName?")
			}
		}

	}

	return nil
}
