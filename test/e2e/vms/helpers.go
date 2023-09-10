package vms

import (
	"fmt"
	"github.com/helloyi/go-sshclient"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/status_codes"
	"go-deploy/test/e2e"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func waitForVmCreated(t *testing.T, id string, callback func(*body.VmRead) bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := e2e.DoGetRequest(t, "/vms/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var vmRead body.VmRead
		err := e2e.ReadResponseBody(t, resp, &vmRead)
		if err != nil {
			continue
		}

		if vmRead.Status == status_codes.GetMsg(status_codes.ResourceRunning) {
			finished := callback(&vmRead)
			if finished {
				break
			}
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "vm did not start in time") {
			assert.FailNow(t, "vm did not start in time")
			break
		}
	}
}

func waitForVmDeleted(t *testing.T, id string, callback func() bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := e2e.DoGetRequest(t, "/vms/"+id)
		if resp.StatusCode == http.StatusNotFound {
			if callback() {
				break
			}
			break
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "vm did not start in time") {
			assert.FailNow(t, "vm did not start in time")
			break
		}
	}
}

func withSshPublicKey(t *testing.T) string {
	content, err := os.ReadFile("../../ssh/id_rsa.pub")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return strings.TrimSpace(string(content))
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

	client, err := sshclient.DialWithKey(fmt.Sprintf("%s:%s", address, port), user, "../../ssh/id_rsa")
	if err != nil {
		return false
	}

	err = client.Close()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	return true
}
