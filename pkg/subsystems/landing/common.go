package landing

import (
	"encoding/json"
	"fmt"
	"go-deploy/utils/requestutils"
	"net/http"
)

func (client *Client) doRequest(method string, relativePath string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", client.url, relativePath)
	return requestutils.DoRequestBearer(method, fullURL, nil, nil, client.jwt.AccessToken)
}

func (client *Client) doQueryRequest(method string, relativePath string, params map[string]string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", client.url, relativePath)
	return requestutils.DoRequestBearer(method, fullURL, nil, params, client.jwt.AccessToken)
}

func (client *Client) doJSONRequest(method string, relativePath string, requestBody interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s%s", client.url, relativePath)
	return requestutils.DoRequestBearer(method, fullURL, jsonBody, nil, client.jwt.AccessToken)
}
