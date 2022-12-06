package npm

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/utils/requestutils"
	"log"
)

func deleteProxyHost(name string, token string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create proxy host. details: %s", err)
	}

	proxyHost, err := fetchProxyHost(name, token)
	if err != nil {
		return makeError(err)
	}

	if proxyHost.ID == 0 {
		return nil
	}

	deleteProxyHostUrl := fmt.Sprintf("%s/nginx/proxy-hosts/%d", conf.Env.Npm.Url, proxyHost.ID)
	res, err := requestutils.DoRequestBearer("DELETE", deleteProxyHostUrl, nil, token)
	if err != nil {
		return makeError(err)
	}

	defer requestutils.CloseBody(res.Body)

	body, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return makeApiError(body, makeError)
	}

	return nil
}

func Delete(name string) error {
	log.Println("deleting npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete npm setup for project %s. details: %s", name, err)
	}

	token, err := createToken()
	if err != nil {
		return makeError(err)
	}

	err = deleteProxyHost(name, token)
	if err != nil {
		return makeError(err)
	}

	return nil
}
