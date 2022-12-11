package requestutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func IsGoodStatusCode(code int) bool {
	return int(code/100) == 2
}

func IsUserError(code int) bool {
	return int(code/100) == 4
}

func IsInternalError(code int) bool {
	return int(code/100) == 5
}

func setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func setBearerTokenHeaders(req *http.Request, token string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func setBasicAuthHeaders(req *http.Request, username string, password string) {
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)
}

func doRequestInternal(req *http.Request) (*http.Response, error) {
	// do request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to do http request. details: %s", err)
		return nil, err
	}

	// check if we received anything
	if res.Body == nil {
		err = fmt.Errorf("failed to open response. details: no body")
		return nil, err
	}

	return res, nil
}

func DoRequest(method string, url string, requestBody []byte) (*http.Response, error) {
	// prepare request
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	setHeaders(req)
	return doRequestInternal(req)
}

func ParseBody[T any](closer io.ReadCloser, out *T) error {
	body, err := ReadBody(closer)
	if err != nil {
		return err
	}
	defer CloseBody(closer)

	err = ParseJson(body, out)
	if err != nil {
		return err
	}

	return nil
}

func DoRequestBearer(method string, url string, requestBody []byte, token string) (*http.Response, error) {
	// prepare request
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	setBearerTokenHeaders(req, token)
	return doRequestInternal(req)
}

func DoRequestBasicAuth(method string, url string, requestBody []byte, username string, password string) (*http.Response, error) {
	// prepare request
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	setBasicAuthHeaders(req, username, password)
	return doRequestInternal(req)
}

func CloseBody(Body io.ReadCloser) {
	err := Body.Close()
	if err != nil {
		log.Println("failed to close response body. details: ", err)
	}
}

func ReadBody(responseBody io.ReadCloser) ([]byte, error) {
	// read body
	body, err := io.ReadAll(responseBody)
	if err != nil {
		err = fmt.Errorf("failed to read response body. details: %s", err)
		return nil, err
	}
	return body, nil
}

func ParseJson[T any](data []byte, out *T) error {
	err := json.Unmarshal(data, out)
	if err != nil {
		err = fmt.Errorf("failed to parse json data. details: %s", err)
		return err
	}
	return nil
}
