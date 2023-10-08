package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"log"
)

type SsDeleterType[IdType any] struct {
	id         *string
	resourceID *IdType
	dbKey      string

	dbFunc     func(string, string, interface{}) error
	deleteFunc func(IdType) error
}

func SsDeleter[IdType any](deleteFunc func(IdType) error) *SsDeleterType[IdType] {
	return &SsDeleterType[IdType]{
		deleteFunc: deleteFunc,
		dbFunc:     deploymentModel.New().UpdateSubsystemByID_test,
	}
}

func (rc *SsDeleterType[IdType]) WithID(id string) *SsDeleterType[IdType] {
	rc.id = &id
	return rc
}

func (rc *SsDeleterType[IdType]) WithResourceID(resourceID IdType) *SsDeleterType[IdType] {
	rc.resourceID = &resourceID
	return rc
}

func (rc *SsDeleterType[IdType]) WithDbFunc(dbFunc func(string, string, interface{}) error) *SsDeleterType[IdType] {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsDeleterType[IdType]) WithDbKey(dbKey string) *SsDeleterType[IdType] {
	rc.dbKey = dbKey
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

	if rc.dbKey == "" {
		log.Println("no db key provided for subsystem creation. did you forget to call WithDbKey?")
	} else {
		if rc.id != nil {
			return rc.dbFunc(*rc.id, rc.dbKey, nil)
		} else {
			log.Println("no id or name provided for subsystem update. did you forget to call WithID?")
		}
	}

	return nil
}
