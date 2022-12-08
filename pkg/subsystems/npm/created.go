package npm

import (
	"fmt"
)

func createdProxyHost(name string, token string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check created npm proxy host. details: %s", err)
	}

	proxyHost, err := fetchProxyHost(name, token)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHost.ID != 0, nil
}

func Created(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if npm setup is created for deployment %s. details: %s", name, err)
	}

	token, err := createToken()
	if err != nil {
		return false, makeError(err)
	}

	proxyHostCreated, err := createdProxyHost(name, token)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHostCreated, nil
}
