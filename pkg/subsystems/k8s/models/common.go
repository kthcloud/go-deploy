package models

import (
	"errors"
	"github.com/google/uuid"
)

func GetIdAndName(k8sName string) ([]string, error) {
	uuidLen := len(uuid.New().String())
	if len(k8sName) < uuidLen {
		return nil, errors.New("could not parse id and name")
	}

	id := k8sName[len(k8sName)-uuidLen:]
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("could not parse id and name")
	}

	name := k8sName[:len(k8sName)-uuidLen-1]

	return []string{id, name}, nil
}
