package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/pkg/sys"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	AdminUserID   = "955f0f87-37fd-4792-90eb-9bf6989e698a"
	PowerUserID   = "955f0f87-37fd-4792-90eb-9bf6989e698b"
	DefaultUserID = "955f0f87-37fd-4792-90eb-9bf6989e698c"
	TestDomain    = "test-deploy.saffronbun.com"
	CheckInterval = 1 * time.Second
	MaxChecks     = 900 // 900 * CheckInterval (1) seconds = 15 minutes

	// V2TestsEnabled flag is used temporarily to enable V2 tests until V2 zone is fully operational
	V2TestsEnabled = true
)

func FetchUntil(t *testing.T, subPath string, callback func(*http.Response) bool) {
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
		if loops > MaxChecks {
			assert.FailNow(t, "model fetch timeout")
		}
	}
}

func GenName(base ...string) string {
	if len(base) == 0 {
		return "e2e-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
	}

	return "e2e-" + strings.ReplaceAll(base[0], " ", "-") + "-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
}

func StrPtr(s string) *string {
	return &s
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
	return protocol + "://localhost:8080" + subPath
}

func DoGetRequest(t *testing.T, subPath string, userID ...string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", CreateServerURL(subPath), nil)
	assert.NoError(t, err)

	// Set go-deploy-test-user header
	effectiveUser := AdminUserID
	if len(userID) > 0 {
		effectiveUser = userID[0]
	}
	req.Header.Set("go-deploy-test-user", effectiveUser)

	return doRequest(t, req)
}

func DoPostRequest(t *testing.T, subPath string, body interface{}, userID ...string) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBody)

	req, err := http.NewRequest("POST", CreateServerURL(subPath), bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	// Set go-deploy-test-user header
	effectiveUser := AdminUserID
	if len(userID) > 0 {
		effectiveUser = userID[0]
	}
	req.Header.Set("go-deploy-test-user", effectiveUser)

	return doRequest(t, req)
}

func DoDeleteRequest(t *testing.T, subPath string, userID ...string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("DELETE", CreateServerURL(subPath), nil)
	assert.NoError(t, err)

	// Set go-deploy-test-user header
	effectiveUser := AdminUserID
	if len(userID) > 0 {
		effectiveUser = userID[0]
	}
	req.Header.Set("go-deploy-test-user", effectiveUser)

	return doRequest(t, req)
}

func doRequest(t *testing.T, req *http.Request) *http.Response {
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

func ReadRawResponseBody(t *testing.T, resp *http.Response) []byte {
	t.Cleanup(func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	})

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to read response body: %s", err.Error()))
	}

	return raw
}

func parseRawBody(body []byte, parsedBody interface{}) error {
	err := json.Unmarshal(body, parsedBody)
	if err != nil {
		return fmt.Errorf("failed to parse body: %w", err)
	}

	return nil
}

func Parse[okType any](t *testing.T, resp *http.Response) okType {
	if resp.StatusCode > 299 {
		rawBody := ReadRawResponseBody(t, resp)
		empty := new(okType)

		var bindingError body.BindingError
		err := parseRawBody(rawBody, &bindingError)

		// Check if it was a binding error
		if err != nil || len(bindingError.ValidationErrors) > 0 {
			for _, fieldErrors := range bindingError.ValidationErrors {
				for _, fieldError := range fieldErrors {
					assert.Fail(t, fmt.Sprintf("binding error: %s", fieldError))
				}
			}

			assert.FailNow(t, fmt.Sprintf("error that was not go-deploy binding error (path: %s status code: %d)", resp.Request.URL.Path, resp.StatusCode))
			return *empty
		}

		// Otherwise parse as ordinary error response
		var errResp sys.ErrorResponse
		err = parseRawBody(rawBody, &errResp)
		if err != nil {
			assert.FailNow(t, fmt.Sprintf("failed to parse error response (status code: %d): %s", resp.StatusCode, err.Error()))
			return *empty
		}

		if len(errResp.Errors) == 0 {
			assert.FailNow(t, fmt.Sprintf("error response has no errors (status code: %d)", resp.StatusCode))
			return *empty
		}

		assert.FailNow(t, fmt.Sprintf("error response has errors (status code: %d): %s", resp.StatusCode, errResp.Errors[0].Msg))
		return *empty
	}

	var okResp okType
	err := ReadResponseBody(t, resp, &okResp)
	assert.NoError(t, err)

	return okResp
}
