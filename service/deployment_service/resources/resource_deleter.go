package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"log"
)

type SsDeleterType[IdType any] struct {
	id         *string
	name       *string
	resourceID *IdType
	dbKey      string
	deleteFunc func(IdType) error
}

func SsDeleter[IdType any](deleteFunc func(IdType) error) *SsDeleterType[IdType] {
	return &SsDeleterType[IdType]{
		deleteFunc: deleteFunc,
	}
}

func (rc *SsDeleterType[IdType]) WithID(id string) *SsDeleterType[IdType] {
	rc.id = &id
	return rc
}

func (rc *SsDeleterType[IdType]) WithName(name string) *SsDeleterType[IdType] {
	rc.name = &name
	return rc
}

func (rc *SsDeleterType[IdType]) WithResourceID(resourceID IdType) *SsDeleterType[IdType] {
	rc.resourceID = &resourceID
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
			return deploymentModel.New().UpdateSubsystemByID_test(*rc.id, rc.dbKey, nil)
		} else if rc.name != nil {
			return deploymentModel.New().UpdateSubsystemByName_test(*rc.name, rc.dbKey, nil)
		} else {
			log.Println("no id or name provided for subsystem update. did you forget to call WithID or WithName?")
		}
	}

	return nil
}
