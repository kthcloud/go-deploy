package resources

import (
	"log"
)

type SsDeleterType[IdType any] struct {
	resourceID *IdType

	dbFunc     func(interface{}) error
	deleteFunc func(IdType) error
}

func SsDeleter[IdType any](deleteFunc func(IdType) error) *SsDeleterType[IdType] {
	return &SsDeleterType[IdType]{
		deleteFunc: deleteFunc,
	}
}

func (rc *SsDeleterType[IdType]) WithResourceID(resourceID IdType) *SsDeleterType[IdType] {
	rc.resourceID = &resourceID
	return rc
}

func (rc *SsDeleterType[IdType]) WithDbFunc(dbFunc func(interface{}) error) *SsDeleterType[IdType] {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsDeleterType[IdType]) Exec() error {
	if rc.resourceID == nil {
		log.Println("no resource id provided for subsystem deletion. did you forget to call WithResourceID?")
	} else if rc.deleteFunc == nil {
		log.Println("no delete function provided for subsystem deletion. did you forget to specify it in the constructor?")
	} else {
		err := rc.deleteFunc(*rc.resourceID)
		if err != nil {
			return err
		}
	}

	if rc.dbFunc == nil {
		log.Println("no db func provided for subsystem deletion. did you forget to call WithDbFunc?")
		return nil
	}

	return rc.dbFunc(nil)
}
