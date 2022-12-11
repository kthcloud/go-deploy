package npm

import (
	"encoding/json"
	"fmt"
	"go-deploy/utils/requestutils"
	"net/http"
)

func (client *Client) doRequest(method string, relativePath string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", client.apiUrl, relativePath)
	return requestutils.DoRequestBearer(method, fullURL, nil, client.token)
}

func (client *Client) doJSONRequest(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", client.apiUrl, relativePath)
	return requestutils.DoRequestBearer(method, fullURL, jsonBody, client.token)
}

func (client *Client) doJSONRequestUnauthorized(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", client.apiUrl, relativePath)
	return requestutils.DoRequest(method, fullURL, jsonBody)
}
