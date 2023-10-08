package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"log"
)

type SsUpdaterType[T service.SsResource] struct {
	id     *string
	public T

	dbKey  string
	dbFunc func(string, string, interface{}) error

	updateFunc func(T) (T, error)
}

func SsUpdater[T service.SsResource](updateFunc func(T) (T, error)) *SsUpdaterType[T] {
	return &SsUpdaterType[T]{
		dbFunc:     deploymentModel.New().UpdateSubsystemByID_test,
		updateFunc: updateFunc,
	}
}

func (rc *SsUpdaterType[T]) WithID(id string) *SsUpdaterType[T] {
	rc.id = &id
	return rc
}

func (rc *SsUpdaterType[T]) WithDbKey(dbKey string) *SsUpdaterType[T] {
	rc.dbKey = dbKey
	return rc
}

func (rc *SsUpdaterType[T]) WithDbFunc(dbFunc func(string, string, interface{}) error) *SsUpdaterType[T] {
	rc.dbFunc = dbFunc
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
				return rc.dbFunc(*rc.id, rc.dbKey, resource)
			} else {
				log.Println("no id or name provided for subsystem update. did you forget to call WithID?")
			}
		}
	}

	return nil
}
