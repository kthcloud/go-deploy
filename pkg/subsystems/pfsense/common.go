package pfsense

import (
	"encoding/json"
	"fmt"
	"go-deploy/pkg/subsystems/pfsense/models"
	"go-deploy/utils/requestutils"
	"io"
	"net/http"
)

func (client *Client) doRequest(method string, relativePath string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", client.apiUrl, relativePath)
	return requestutils.DoRequestBasicAuth(method, fullURL, nil, client.username, client.password)
}

func (client *Client) doJSONRequest(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", client.apiUrl, relativePath)
	return requestutils.DoRequestBasicAuth(method, fullURL, jsonBody, client.username, client.password)
}

func ParseResponse(response *http.Response) (*models.Response, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to parse pfsense response. details: %s", err.Error())
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer requestutils.CloseBody(response.Body)

	var pfSenseResponse models.Response
	err = requestutils.ParseJson(body, &pfSenseResponse)
	if err != nil {
		return nil, makeError(fmt.Errorf("unknown error %d. details: %s", response.StatusCode, err))
	}

	return &pfSenseResponse, nil
}
