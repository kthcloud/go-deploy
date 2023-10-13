package resources

import (
	"log"
)

type SsPlaceholderCreatorType struct {
	dbFunc func(interface{}) error
}

func SsPlaceholderCreator() *SsPlaceholderCreatorType {
	return &SsPlaceholderCreatorType{}
}

func (rc *SsPlaceholderCreatorType) WithDbFunc(dbFunc func(interface{}) error) *SsPlaceholderCreatorType {
	rc.dbFunc = dbFunc
	return rc
}

func (rc *SsPlaceholderCreatorType) Exec() error {
	if rc.dbFunc == nil {
		log.Println("no db key provided for subsystem placeholder creation. did you forget to call WithDbKey?")
		return nil
	}

	return rc.dbFunc(true)
}
