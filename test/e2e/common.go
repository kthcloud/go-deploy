package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/sys"
	"io"
	"net/http"
	"reflect"
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
	return protocol + "://localhost:8080/v1" + subPath
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

func ParseRawBody(t *testing.T, body []byte, parsedBody interface{}) {
	err := json.Unmarshal(body, parsedBody)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to parse body: %s", err.Error()))
	}
}

// EqualOrEmpty checks if two lists are equal, where [] == nil
func EqualOrEmpty(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// check if "expected" is a slice, and if so, check how many elements it has
	isSlice := reflect.ValueOf(expected).Kind() == reflect.Slice
	noElements := 0
	if isSlice {
		noElements = reflect.ValueOf(expected).Len()
	}

	if expected == nil || (isSlice && noElements == 0) {
		assert.Empty(t, actual, msgAndArgs)
	} else {
		assert.EqualValues(t, expected, actual, msgAndArgs)
	}
}

func Parse[okType any](t *testing.T, resp *http.Response) okType {
	if resp.StatusCode > 299 {
		rawBody := ReadRawResponseBody(t, resp)

		var bindingError body.BindingError
		ParseRawBody(t, rawBody, &bindingError)

		// Check if it was a binding error
		if len(bindingError.ValidationErrors) > 0 {
			for _, fieldErrors := range bindingError.ValidationErrors {
				for _, fieldError := range fieldErrors {
					assert.Fail(t, fmt.Sprintf("binding error: %s", fieldError))
				}
			}

			assert.FailNow(t, fmt.Sprintf("binding error (status code: %d)", resp.StatusCode))
		}

		// Otherwise parse as ordinary error response
		var errResp sys.ErrorResponse
		ParseRawBody(t, rawBody, &errResp)

		empty := new(okType)

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
