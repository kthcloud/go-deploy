package e2e

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	TestUserID    = "955f0f87-37fd-4792-90eb-9bf6989e698e"
	TestDomain    = "test-deploy.saffronbun.com"
	CheckInterval = 1 * time.Second
)

func fetchUntil(t *testing.T, subPath string, callback func(*http.Response) bool) {
	loops := 0
	for {
		time.Sleep(CheckInterval)

		resp := DoGetRequest(t, subPath)
		if resp.StatusCode == http.StatusNotFound {
			if callback == nil || callback(resp) {
				break
			}
		}

		if callback == nil || callback(resp) {
			break
		}

		loops++
		if !assert.LessOrEqual(t, loops, 600, "resource fetch timeout") {
			break
		}
	}
}

func GenName(base string) string {
	return base + "-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
}

func DoPlainGetRequest(t *testing.T, path string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", path, nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func CreateServerURL(subPath string) string {
	return CreateServerUrlWithProtocol("http", subPath)
}

func CreateServerUrlWithProtocol(protocol, subPath string) string {
	return protocol + "://localhost:8080/v1" + subPath
}

func DoGetRequest(t *testing.T, subPath string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", CreateServerURL(subPath), nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func DoPostRequest(t *testing.T, subPath string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBody)

	req, err := http.NewRequest("POST", CreateServerURL(subPath), bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func DoDeleteRequest(t *testing.T, subPath string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("DELETE", CreateServerURL(subPath), nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func ReadResponseBody(t *testing.T, resp *http.Response, body interface{}) error {
	t.Cleanup(func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	})

	return json.NewDecoder(resp.Body).Decode(body)
}
