package npm

import (
	"fmt"
)

func deletedProxyHost(name string, token string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check deleted proxy host. details: %s", err)
	}

	proxyHost, err := fetchProxyHost(name, token)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHost.ID == 0, nil
}

func Deleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is deleted for project %s. details: %s", name, err)
	}

	token, err := createToken()
	if err != nil {
		return false, makeError(err)

	}

	proxyHostDeleted, err := deletedProxyHost(name, token)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostDeleted, nil

}
