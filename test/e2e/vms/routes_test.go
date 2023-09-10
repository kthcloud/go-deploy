package vms

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/test/e2e"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestGetVms(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/vms")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateVm(t *testing.T) {
	publicKey := withSshPublicKey(t)

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
		Zone:     nil,
	}

	resp := e2e.DoPostRequest(t, "/vms", requestBody)
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "vm was not created") {
		assert.FailNow(t, "vm was not created")
	}

	var vmCreated body.VmCreated
	err := e2e.ReadResponseBody(t, resp, &vmCreated)
	assert.NoError(t, err, "vm was not created")

	t.Cleanup(func() {
		resp = e2e.DoDeleteRequest(t, "/vms/"+vmCreated.ID)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			assert.FailNow(t, "resource was not deleted")
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
