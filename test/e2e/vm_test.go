package e2e

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/status_codes"
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

		resp := doGetRequest(t, "/vms/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var vmRead body.VmRead
		err := readResponseBody(t, resp, &vmRead)
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

		resp := doGetRequest(t, "/vms/"+id)
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
	content, err := os.ReadFile("../ssh/id_rsa.pub")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return strings.TrimSpace(string(content))
}

func TestGetVms(t *testing.T) {
	setup(t)
	withServer(t)

	resp := doGetRequest(t, "/vms")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateVm(t *testing.T) {
	setup(t)
	withServer(t)
	publicKey := withSshPublicKey(t)

	zone := conf.Env.CS.GetZoneByName("Flemingsberg")
	if zone == nil {
		t.Fatal("zone not found")
	}

	requestBody := body.VmCreate{
		Name:         "e2e-" + strings.ReplaceAll(uuid.NewString()[:10], "-", ""),
		SshPublicKey: publicKey,
		Ports: []body.Port{
			{
				Name:     "e2e-test",
				Port:     100,
				Protocol: "tcp",
			},
		},
		CpuCores: 2,
		RAM:      2,
		DiskSize: 20,
		Zone:     &zone.ID,
	}

	resp := doPostRequest(t, "/vms", requestBody)
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "vm was not created") {
		assert.FailNow(t, "vm was not created")
	}

	var vmCreated body.VmCreated
	err := readResponseBody(t, resp, &vmCreated)
	assert.NoError(t, err, "vm was not created")

	t.Cleanup(func() {
		resp = doDeleteRequest(t, "/vms/"+vmCreated.ID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "vm was not deleted")
		if !assert.Equal(t, http.StatusOK, resp.StatusCode, "vm was not deleted") {
			assert.FailNow(t, "vm was not deleted")
		}

		waitForVmDeleted(t, vmCreated.ID, func() bool {
			return true
		})
	})

	waitForVmCreated(t, vmCreated.ID, func(vmRead *body.VmRead) bool {
		//make sure it is accessible
		if vmRead.ConnectionString != nil {
			return checkUpVM(t, *vmRead.ConnectionString)
		}
		return false
	})
}
