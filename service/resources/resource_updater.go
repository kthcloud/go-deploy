package resources

import (
	"go-deploy/service"
	"log"
)

type SsUpdaterType[T service.SsResource] struct {
	public T

	dbFunc     func(interface{}) error
	updateFunc func(T) (T, error)
}

func SsUpdater[T service.SsResource](updateFunc func(T) (T, error)) *SsUpdaterType[T] {
	return &SsUpdaterType[T]{
		updateFunc: updateFunc,
	}
}

func (rc *SsUpdaterType[T]) WithDbFunc(dbFunc func(interface{}) error) *SsUpdaterType[T] {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsUpdaterType[T]) WithPublic(public T) *SsUpdaterType[T] {
	rc.public = public
	return rc
}

func (rc *SsUpdaterType[T]) Exec() error {
	if service.Nil(rc.public) {
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

		if service.Nil(resource) {
			log.Println("no resource returned after update. assuming it was deleted")
			return nil
		}

		if rc.dbFunc == nil {
			log.Println("no db func provided for subsystem update. did you forget to call WithDbFunc?")
			return nil
		}

		return rc.dbFunc(resource)
	}

	return nil
}
