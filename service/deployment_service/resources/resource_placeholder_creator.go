package resources

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"log"
)

type SsPlaceholderCreatorType struct {
	id    *string
	name  *string
	dbKey string
}

func SsPlaceholderCreator() *SsPlaceholderCreatorType {
	return &SsPlaceholderCreatorType{}
}

func (rc *SsPlaceholderCreatorType) WithID(id string) *SsPlaceholderCreatorType {
	rc.id = &id
	return rc
}

func (rc *SsPlaceholderCreatorType) WithName(name string) *SsPlaceholderCreatorType {
	rc.name = &name
	return rc
}

func (rc *SsPlaceholderCreatorType) WithDbKey(dbKey string) *SsPlaceholderCreatorType {
	rc.dbKey = dbKey
	return rc
}

func (rc *SsPlaceholderCreatorType) Exec() error {
	if rc.dbKey == "" {
		log.Println("no db key provided for subsystem placeholder creation. did you forget to call WithDbKey?")
		return nil
	}

	fullDbKey := rc.dbKey + ".placeholder"

	if rc.id != nil {
		return deploymentModel.New().UpdateSubsystemByID_test(*rc.id, fullDbKey, true)
	}

	if rc.name != nil {
		return deploymentModel.New().UpdateSubsystemByName_test(*rc.name, fullDbKey, true)
	}

	log.Println("no id or name provided for subsystem placeholder creation. did you forget to call WithID or WithName?")
	return nil
}
