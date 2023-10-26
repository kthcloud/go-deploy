package e2e

import (
	"fmt"
	"github.com/helloyi/go-sshclient"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	status_codes2 "go-deploy/pkg/app/status_codes"
	"net/http"
	"strings"
	"testing"
)

func WaitForVmRunning(t *testing.T, id string, callback func(*body.VmRead) bool) {
	fetchUntil(t, "/vms/"+id, func(resp *http.Response) bool {
		var vmRead body.VmRead
		err := ReadResponseBody(t, resp, &vmRead)
		assert.NoError(t, err, "vm was not fetched")

		if vmRead.Status == status_codes2.GetMsg(status_codes2.ResourceRunning) {
			if callback == nil || callback(&vmRead) {
				return true
			}
		}

		return false
	})
}

func WaitForVmDeleted(t *testing.T, id string, callback func() bool) {
	fetchUntil(t, "/vms/"+id, func(resp *http.Response) bool {
		if resp.StatusCode != http.StatusNotFound {
			return false
		}

		if callback == nil {
			return true
		}

		return callback()
	})
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
	assert.NoError(t, err, "ssh connection was not closed")

	return true
}

func WithVM(t *testing.T, requestBody body.VmCreate) body.VmRead {
	resp := DoPostRequest(t, "/vms", requestBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "deployment was not created")

	var vmCreated body.VmCreated
	err := ReadResponseBody(t, resp, &vmCreated)
	assert.NoError(t, err, "vm was not created")

	t.Cleanup(func() { cleanUpVm(t, vmCreated.ID) })

	WaitForJobFinished(t, vmCreated.JobID, nil)
	WaitForVmRunning(t, vmCreated.ID, func(vmRead *body.VmRead) bool {
		//make sure it is accessible
		if vmRead.ConnectionString != nil {
			return checkUpVM(t, *vmRead.ConnectionString)
		}
		return false
	})

	var vmRead body.VmRead
	readResp := DoGetRequest(t, "/vms/"+vmCreated.ID)
	err = ReadResponseBody(t, readResp, &vmRead)
	assert.NoError(t, err, "vm was not created")

	assert.NotEmpty(t, vmRead.ID)
	assert.Equal(t, requestBody.Name, vmRead.Name)
	assert.Equal(t, requestBody.SshPublicKey, vmRead.SshPublicKey)
	assert.Equal(t, requestBody.CpuCores, vmRead.Specs.CpuCores)
	assert.Equal(t, requestBody.RAM, vmRead.Specs.RAM)
	assert.Equal(t, requestBody.DiskSize, vmRead.Specs.DiskSize)

	if requestBody.Ports == nil {
		assert.Empty(t, vmRead.Ports)
	} else {
		for _, port := range requestBody.Ports {
			found := false
			for _, portRead := range vmRead.Ports {
				if port.Name == portRead.Name {
					assert.Equal(t, port.Port, portRead.Port)
					assert.Equal(t, port.Protocol, portRead.Protocol)
					assert.NotZero(t, portRead.ExternalPort)
					found = true
					break
				}
			}
			assert.True(t, found, "port not found")
		}
	}

	if requestBody.Zone == nil {
		// some zone is set by default
		assert.NotEmpty(t, vmRead.Zone)
	} else {
		assert.Equal(t, requestBody.Zone, vmRead.Zone)
	}

	return vmRead
}

func WithAssumedFailedVM(t *testing.T, requestBody body.VmCreate) {
	resp := DoPostRequest(t, "/vms", requestBody)
	if resp.StatusCode == http.StatusBadRequest {
		return
	}

	var vmCreated body.VmCreated
	err := ReadResponseBody(t, resp, &vmCreated)
	assert.NoError(t, err, "created body was not read")

	t.Cleanup(func() { cleanUpVm(t, vmCreated.ID) })

	assert.FailNow(t, "resource was created but should have failed")
}

func cleanUpVm(t *testing.T, id string) {
	resp := DoDeleteRequest(t, "/vms/"+id)
	if resp.StatusCode == http.StatusNotFound {
		return
	}

	if resp.StatusCode == http.StatusOK {
		var vmDeleted body.VmDeleted
		err := ReadResponseBody(t, resp, &vmDeleted)
		assert.NoError(t, err, "deleted body was not read")
		assert.Equal(t, id, vmDeleted.ID)

		WaitForJobFinished(t, vmDeleted.JobID, nil)
		WaitForVmDeleted(t, vmDeleted.ID, nil)

		return
	}

	assert.FailNow(t, "vm was not deleted")
}
