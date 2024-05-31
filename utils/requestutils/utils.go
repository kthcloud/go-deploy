package requestutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-deploy/utils"
	"io"
	"net/http"
)

// IsGoodStatusCode checks if the status code is a good status code, i.e. 2xx
func IsGoodStatusCode(code int) bool {
	return int(code/100) == 2
}

// IsUserError checks if the status code is a bad request, i.e. 4xx
func IsUserError(code int) bool {
	return int(code/100) == 4
}

// IsInternalError checks if the status code is an internal error, i.e. 5xx
func IsInternalError(code int) bool {
	return int(code/100) == 5
}

// setBearerTokenHeaders sets the bearer token headers for a request
func setBearerTokenHeaders(req *http.Request, token string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

// setBasicAuthHeaders sets the basic auth headers for a request
func setBasicAuthHeaders(req *http.Request, username string, password string) {
	req.SetBasicAuth(username, password)
}

// setJsonHeader sets the json header for a request
func setJsonHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

// doRequestInternal does the actual request
func doRequestInternal(req *http.Request) (*http.Response, error) {
	// do request
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to do http request. details: %w", err)
		return nil, err
	}

	// check if we received anything
	if res.Body == nil {
		err = fmt.Errorf("failed to open response. details: no body")
		return nil, err
	}

	return res, nil
}

// addParams adds params to a request
func addParams(req *http.Request, params map[string]string) {
	values := req.URL.Query()
	for key, value := range params {
		values.Add(key, value)
	}
	req.URL.RawQuery = values.Encode()
}

func DoJsonGetRequest[T any](url string, params map[string]string) (*T, error) {
	res, err := DoRequest("GET", url, nil, params)
	if err != nil {
		return nil, err
	}

	body, err := ParseBody[T](res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// DoRequest is a helper function that does a request with the given method, url, request body and params
func DoRequest(method string, url string, requestBody []byte, params map[string]string) (*http.Response, error) {
	// prepare request
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if params == nil {
		setJsonHeader(req)
	}
	addParams(req, params)
	return doRequestInternal(req)
}

// DoRequestBearer is a helper function that does a request with the given method, url, request body, params and bearer token
func DoRequestBearer(method string, url string, requestBody []byte, params map[string]string, token string) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if params == nil {
		setJsonHeader(req)
	}
	setBearerTokenHeaders(req, token)
	addParams(req, params)
	return doRequestInternal(req)
}

// DoRequestBasicAuth is a helper function that does a request with the given method, url, request body, params and basic auth
func DoRequestBasicAuth(method string, url string, requestBody []byte, params map[string]string, username string, password string) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if params == nil {
		setJsonHeader(req)
	}
	setBasicAuthHeaders(req, username, password)
	addParams(req, params)
	return doRequestInternal(req)
}

// ParseBody is a helper function that parses the body of a response into the given out object
func ParseBody[T any](closer io.ReadCloser) (*T, error) {
	body, err := ReadBody(closer)
	if err != nil {
		return nil, err
	}
	defer CloseBody(closer)

	res, err := ParseJson[T](body)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CloseBody is a helper function that closes the body of a response
func CloseBody(Body io.ReadCloser) {
	err := Body.Close()
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to close response body. details: %w", err))
	}
}

// ReadBody is a helper function that reads the body of a response
func ReadBody(responseBody io.ReadCloser) ([]byte, error) {
	// read body
	body, err := io.ReadAll(responseBody)
	if err != nil {
		err = fmt.Errorf("failed to read response body. details: %w", err)
		return nil, err
	}
	return body, nil
}

// ParseJson is a helper function that parses json data into the given out object
func ParseJson[T any](data []byte) (*T, error) {
	var out T
	err := json.Unmarshal(data, &out)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json data. details: %s", err)
	}
	return &out, nil
}
