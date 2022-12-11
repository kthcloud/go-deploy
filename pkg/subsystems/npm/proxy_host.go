package npm

import (
	"fmt"
	"go-deploy/pkg/subsystems/npm/models"
	"go-deploy/utils/requestutils"
)

func (client *Client) ProxyHostCreated(domainName string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check created npm proxy host. details: %s", err)
	}

	proxyHost, err := client.GetProxyHost(domainName)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHost != nil, nil
}

func (client *Client) ProxyHostDeleted(domainName string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check deleted npm proxy host. details: %s", err)
	}

	proxyHost, err := client.GetProxyHost(domainName)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHost == nil, nil
}

func (client *Client) GetProxyHost(domainName string) (*models.ProxyHost, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch proxy host with domain name %s. details: %s", domainName, err)
	}

	res, err := client.doRequest("GET", "/nginx/proxy-hosts")
	if err != nil {
		return nil, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return nil, makeApiError(res.Body, makeError)
	}

	var proxyHosts []models.ProxyHost
	err = requestutils.ParseBody(res.Body, &proxyHosts)
	if err != nil {
		return nil, makeError(err)
	}

	for _, proxyHost := range proxyHosts {
		for _, name := range proxyHost.DomainNames {
			if name == domainName {
				return &proxyHost, nil
			}
		}
	}

	return nil, nil
}

func (client *Client) CreateProxyHost(domainName string, forwardHost string, port int, certificateId int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create npm proxy host. details: %s", err)
	}

	proxyHostExists, err := client.ProxyHostCreated(domainName)
	if err != nil {
		return makeError(err)
	}

	if proxyHostExists {
		return nil
	}

	requestBody := models.CreateProxyHostBody([]string{domainName}, forwardHost, port, certificateId)
	res, err := client.doJSONRequest("POST", "/nginx/proxy-hosts", requestBody)

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return makeApiError(res.Body, makeError)
	}

	return nil
}

func (client *Client) DeleteProxyHost(domainName string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create npm proxy host. details: %s", err)
	}

	proxyHost, err := client.GetProxyHost(domainName)
	if err != nil {
		return makeError(err)
	}

	if proxyHost == nil {
		return nil
	}

	res, err := client.doRequest("DELETE", fmt.Sprintf("/nginx/proxy-hosts/%d", proxyHost.ID))
	if err != nil {
		return makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return makeApiError(res.Body, makeError)
	}

	return nil
}
