package npm

import (
	"encoding/json"
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/utils/requestutils"
	"log"
)

func getWildcardCertificateID(token string) (int, error) {
	// static
	certificatesUrl := fmt.Sprintf("%s/nginx/certificates", conf.Env.Npm.Url)

	makeError := func(err error) error {
		return fmt.Errorf("failed to get certificates. details: %s", err)
	}

	res, err := requestutils.DoRequestBearer("GET", certificatesUrl, nil, token)
	if err != nil {
		return -1, makeError(err)
	}

	defer requestutils.CloseBody(res.Body)

	body, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return -1, makeError(err)
	}

	// check if good request
	if requestutils.IsGoodStatusCode(res.StatusCode) {
		certificatesParsed := certificatesBody{}
		err = requestutils.ParseJsonBody(body, &certificatesParsed)
		if err != nil {
			return -1, makeError(err)
		}

		wildcard := fmt.Sprintf("*.%s", conf.Env.ParentDomain)

		for _, certificate := range certificatesParsed {
			for _, domainName := range certificate.DomainNames {
				if domainName == wildcard {
					return certificate.ID, nil
				}
			}
		}

		err = fmt.Errorf("failed to find wild card certificate. details: certificate list did not contain a certificate with domain name %s", wildcard)
		return -1, makeError(err)

	} else {
		npmApiErrorRequestParsed := npmApiError{}
		err = requestutils.ParseJsonBody(body, &npmApiErrorRequestParsed)
		if err != nil {
			return -1, makeError(err)
		}

		errorCode := npmApiErrorRequestParsed.Error.Code
		errorMessage := npmApiErrorRequestParsed.Error.Message

		err = fmt.Errorf("failed to fetch certificates (%d). details: %s", errorCode, errorMessage)
		return -1, makeError(err)
	}
}

func createProxyHost(name string, token string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create proxy host. details: %s", err)
	}

	proxyHostCreated, err := createdProxyHost(name, token)
	if err != nil {
		return makeError(err)
	}

	if proxyHostCreated {
		return nil
	}

	id, err := getWildcardCertificateID(token)
	if err != nil {
		return makeError(err)
	}

	requestBody := createProxyHostBody(name, id)
	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return makeError(err)
	}

	proxyHostUrl := fmt.Sprintf("%s/nginx/proxy-hosts", conf.Env.Npm.Url)
	res, err := requestutils.DoRequestBearer("POST", proxyHostUrl, requestBodyJson, token)
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

func Create(name string) error {
	log.Println("creating npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to create npm setup for project %s. details: %s", name, err)
	}

	token, err := createToken()
	if err != nil {
		return makeError(err)
	}

	err = createProxyHost(name, token)
	if err != nil {
		return makeError(err)
	}

	return nil
}
