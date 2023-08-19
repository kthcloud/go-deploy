package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/helloyi/go-sshclient"
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/app"
	"go-deploy/pkg/conf"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func setup(t *testing.T) {

	requiredEnvs := []string{
		"DEPLOY_CONFIG_FILE",
	}

	for _, env := range requiredEnvs {
		_, result := os.LookupEnv(env)
		if !result {
			t.Fatalf("%s must be set for acceptance test", env)
		}
	}

	_, result := os.LookupEnv("DEPLOY_CONFIG_FILE")
	if result {
		conf.SetupEnvironment()
	}
}

func withServer(t *testing.T) *http.Server {
	t.Helper()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	httpServer := app.Start(ctx, nil)
	t.Cleanup(func() {
		cancel()
		app.Stop(httpServer)
	})

	time.Sleep(3 * time.Second)

	return httpServer
}

func doGetRequest(t *testing.T, subPath string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", "http://localhost:8080/v1"+subPath, nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func doPostRequest(t *testing.T, subPath string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBody)

	req, err := http.NewRequest("POST", "http://localhost:8080/v1"+subPath, bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func doDeleteRequest(t *testing.T, subPath string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("DELETE", "http://localhost:8080/v1"+subPath, nil)
	assert.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}

func checkUpDeployment(t *testing.T, url string) bool {
	t.Helper()

	resp, err := http.Get(url)
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	return false
}

func checkUpVM(t *testing.T, connectionString string) bool {
	t.Helper()

	// ssh user@address -p port
	connectionStringParts := strings.Split(connectionString, " ")
	assert.Len(t, connectionStringParts, 4)

	addrParts := strings.Split(connectionStringParts[1], "@")
	assert.Len(t, addrParts, 2)

	user := addrParts[0]
	address := addrParts[1]
	port := connectionStringParts[3]

	client, err := sshclient.DialWithKey(fmt.Sprintf("%s:%s", address, port), user, "../ssh/id_rsa")
	if err != nil {
		return false
	}

	err = client.Close()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	return true
}

func readResponseBody(t *testing.T, resp *http.Response, body interface{}) error {
	t.Cleanup(func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	})

	return json.NewDecoder(resp.Body).Decode(body)
}
