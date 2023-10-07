package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"log"
)

type SsUpdaterType[T service.SsResource] struct {
	id         *string
	name       *string
	public     T
	dbKey      string
	updateFunc func(T) (T, error)
}

func SsUpdater[T service.SsResource](updateFunc func(T) (T, error)) *SsUpdaterType[T] {
	return &SsUpdaterType[T]{
		updateFunc: updateFunc,
	}
}

func (rc *SsUpdaterType[T]) WithID(id string) *SsUpdaterType[T] {
	rc.id = &id
	return rc
}

func (rc *SsUpdaterType[T]) WithName(name string) *SsUpdaterType[T] {
	rc.name = &name
	return rc
}

func (rc *SsUpdaterType[T]) WithDbKey(dbKey string) *SsUpdaterType[T] {
	rc.dbKey = dbKey
	return rc
}

func (rc *SsUpdaterType[T]) WithPublic(public T) *SsUpdaterType[T] {
	rc.public = public
	return rc
}

func (rc *SsUpdaterType[T]) Exec() error {
	if rc.public == nil {
		log.Println("no public resource provided for subsystem update. did you forget to call WithPublic?")
		return nil
	}

	if service.NotCreated(rc.public) {
		var resource T
		if rc.updateFunc == nil {
			log.Println("no update function provided for subsystem update. did you forget to specify it in the constructor?")
			resource = rc.public
		} else {
			var err error
			resource, err = rc.updateFunc(rc.public)
			if err != nil {
				return err
			}
		}

		if resource == nil {
			log.Println("no resource returned after update. assuming it was deleted")
		} else if rc.dbKey == "" {
			log.Println("no db key provided for subsystem creation. did you forget to call WithDbKey?")
		} else {
			if rc.id != nil {
				return deploymentModel.New().UpdateSubsystemByID_test(*rc.id, rc.dbKey, resource)
			} else if rc.name != nil {
				return deploymentModel.New().UpdateSubsystemByName_test(*rc.name, rc.dbKey, resource)
			} else {
				log.Println("no id or name provided for subsystem update. did you forget to call WithID or WithName?")
			}
		}
	}

	return nil
}
