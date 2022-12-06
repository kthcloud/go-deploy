package npm

import (
	"deploy-api-go/pkg/conf"
	"deploy-api-go/utils/requestutils"
	"encoding/json"
	"fmt"
)

func createToken() (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create token. details: %s", err)
	}

	// statics
	tokenUrl := fmt.Sprintf("%s/tokens", conf.Env.Npm.Url)

	// prepare for request
	requestBodyJson, err := json.Marshal(tokenRequestBody{
		Identity: conf.Env.Npm.Identity,
		Secret:   conf.Env.Npm.Secret,
	})

	res, err := requestutils.DoRequest("POST", tokenUrl, requestBodyJson)
	if err != nil {
		return "", makeError(err)
	}

	defer requestutils.CloseBody(res.Body)

	// read body
	body, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return "", makeError(err)
	}

	// check if good request
	if requestutils.IsGoodStatusCode(res.StatusCode) {
		tokenParsed := tokenBody{}
		err = requestutils.ParseJsonBody(body, &tokenParsed)
		if err != nil {
			return "", makeError(err)
		}
		return tokenParsed.Token, nil
	} else {
		npmApiErrorRequestParsed := npmApiError{}
		err = requestutils.ParseJsonBody(body, &npmApiErrorRequestParsed)
		if err != nil {
			return "", makeError(err)
		}

		resCode := npmApiErrorRequestParsed.Error.Code
		resMessage := npmApiErrorRequestParsed.Error.Message

		err = fmt.Errorf("failed to fetch token (%d). details: %s", resCode, resMessage)
		return "", makeError(err)
	}

}

func fetchProxyHost(name string, token string) (listProxyHostResponseBody, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch proxy host by name. details: %s", err)
	}

	proxyHostUrl := fmt.Sprintf("%s/nginx/proxy-hosts", conf.Env.Npm.Url)

	res, err := requestutils.DoRequestBearer("GET", proxyHostUrl, nil, token)
	if err != nil {
		return listProxyHostResponseBody{}, makeError(err)
	}

	defer requestutils.CloseBody(res.Body)

	body, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return listProxyHostResponseBody{}, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return listProxyHostResponseBody{}, makeApiError(body, makeError)
	}

	listProxyHost := listProxyHostsResponseBody{}
	err = requestutils.ParseJsonBody(body, &listProxyHost)
	if err != nil {
		return listProxyHostResponseBody{}, makeError(err)
	}

	searchFor := getFqdn(name)
	for _, proxyHost := range listProxyHost {
		for _, domainName := range proxyHost.DomainNames {
			if domainName == searchFor {
				return proxyHost, nil
			}
		}
	}

	return listProxyHostResponseBody{}, nil
}
