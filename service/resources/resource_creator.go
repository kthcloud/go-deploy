package resources

import (
	"go-deploy/pkg/subsystems"
	"log"
)

type SsCreatorType[T subsystems.SsResource] struct {
	name   *string
	public T

	dbFunc     func(interface{}) error
	createFunc func(T) (T, error)
}

func SsCreator[T subsystems.SsResource](createFunc func(T) (T, error)) *SsCreatorType[T] {
	return &SsCreatorType[T]{
		createFunc: createFunc,
	}
}

func (rc *SsCreatorType[T]) WithDbFunc(dbFunc func(interface{}) error) *SsCreatorType[T] {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsCreatorType[T]) WithPublic(public T) *SsCreatorType[T] {
	rc.public = public
	return rc
}

func (rc *SsCreatorType[T]) Exec() error {
	if subsystems.Nil(rc.public) {
		log.Println("no public resource provided for subsystem creation. assuming it failed to create")
		return nil
	}

	if subsystems.NotCreated(rc.public) {
		var resource T
		if rc.createFunc == nil {
			log.Println("no create function provided for subsystem creation. did you forget to specify it in the constructor?")
			return nil
		} else {
			var err error
			resource, err = rc.createFunc(rc.public)
			if err != nil {
				return err
			}
		}

		if subsystems.Nil(resource) {
			log.Println("no resource returned after creation. assuming it failed to create or was skipped")
			return nil
		}

		if rc.dbFunc == nil {
			log.Println("no db func provided for subsystem creation. did you forget to call WithDbFunc?")
			return nil
		}

		subsystems.CopyValue(resource, rc.public)
		return rc.dbFunc(resource)
	}

	return nil
}
