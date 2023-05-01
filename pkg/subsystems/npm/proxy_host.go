package npm

import (
	"fmt"
	"go-deploy/pkg/subsystems/npm/models"
	"go-deploy/utils/requestutils"
	"net/http"
)

func (client *Client) ProxyHostCreated(id int) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check created proxy host. details: %s", err)
	}

	if id == 0 {
		return false, fmt.Errorf("id required")
	}

	proxyHost, err := client.ReadProxyHost(id)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHost != nil, nil
}

func (client *Client) ProxyHostDeleted(id int) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check deleted proxy host. details: %s", err)
	}

	if id == 0 {
		return false, fmt.Errorf("id required")
	}

	proxyHost, err := client.ReadProxyHost(id)
	if err != nil {
		return false, makeError(err)
	}

	return proxyHost == nil, nil
}

func (client *Client) ReadProxyHost(id int) (*models.ProxyHostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read proxy host %d. details: %s", id, err)
	}

	if id == 0 {
		return nil, fmt.Errorf("id required")
	}

	res, err := client.doRequest("GET", fmt.Sprintf("/nginx/proxy-hosts/%d", id))
	if err != nil {
		return nil, makeError(err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return nil, makeApiError(res.Body, makeError)
	}

	var proxyHost models.ProxyHostRead
	err = requestutils.ParseBody(res.Body, &proxyHost)
	if err != nil {
		return nil, makeError(err)
	}

	public := models.CreateProxyHostPublicFromReadBody(&proxyHost)

	return public, nil
}

func (client *Client) ReadProxyHostByDomainName(domainName string) (*models.ProxyHostPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read proxy host with domain name %s. details: %s", domainName, err)
	}

	res, err := client.doRequest("GET", "/nginx/proxy-hosts/")
	if err != nil {
		return nil, makeError(err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return nil, makeApiError(res.Body, makeError)
	}

	var proxyHosts []models.ProxyHostRead
	err = requestutils.ParseBody(res.Body, &proxyHosts)
	if err != nil {
		return nil, makeError(err)
	}

	for _, proxyHost := range proxyHosts {
		for _, name := range proxyHost.DomainNames {
			if name == domainName {
				public := models.CreateProxyHostPublicFromReadBody(&proxyHost)
				return public, nil
			}
		}
	}

	return nil, nil
}

func (client *Client) CreateProxyHost(public *models.ProxyHostPublic) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create proxy host. details: %s", err)
	}

	if len(public.DomainNames) == 0 {
		return 0, makeError(fmt.Errorf("no domain names supplied"))
	}

	result, err := client.ReadProxyHostByDomainName(public.DomainNames[0])
	if err != nil {
		return 0, makeError(err)
	}

	if result != nil {
		return result.ID, nil
	}

	requestBody := models.CreateProxyHostCreateBody(public)
	res, err := client.doJsonRequest("POST", "/nginx/proxy-hosts", requestBody)
	if err != nil {
		return 0, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return 0, makeApiError(res.Body, makeError)
	}

	var proxyHostCreated models.ProxyHostCreated
	err = requestutils.ParseBody(res.Body, &proxyHostCreated)
	if err != nil {
		return 0, makeError(err)
	}

	return proxyHostCreated.ID, nil
}

func (client *Client) DeleteProxyHost(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete proxy host. details: %s", err)
	}

	if id == 0 {
		return fmt.Errorf("id required")
	}

	res, err := client.doRequest("DELETE", fmt.Sprintf("/nginx/proxy-hosts/%d", id))
	if err != nil {
		return makeError(err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return makeApiError(res.Body, makeError)
	}

	return nil
}

func (client *Client) UpdateProxyHost(public *models.ProxyHostPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update proxy host. details: %s", err)
	}

	if public.ID == 0 {
		return fmt.Errorf("id required in public body")
	}

	requestBody := models.CreateProxyHostUpdateBody(public)
	res, err := client.doJsonRequest("PUT", fmt.Sprintf("/nginx/proxy-hosts/%d", public.ID), requestBody)
	if err != nil {
		return makeError(err)
	}

	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return makeApiError(res.Body, makeError)
	}

	return nil
}
