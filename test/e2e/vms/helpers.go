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

func waitForJobFinished(t *testing.T, id string, callback func(*body.JobRead) bool) {
	loops := 0
	for {
		time.Sleep(10 * time.Second)

		resp := e2e.DoGetRequest(t, "/jobs/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jobRead body.JobRead
		err := e2e.ReadResponseBody(t, resp, &jobRead)
		if err != nil {
			continue
		}

		if jobRead.Status == status_codes.GetMsg(status_codes.JobFinished) {
			finished := callback(&jobRead)
			if finished {
				break
			}
		}

		if jobRead.Status == status_codes.GetMsg(status_codes.JobTerminated) {
			finished := callback(&jobRead)
			if finished {
				break
			}
		}

		loops++
		if !assert.LessOrEqual(t, loops, 30, "job did not finish in time") {
			assert.FailNow(t, "job did not finish in time")
			break
		}
	}
}

func waitForVmRunning(t *testing.T, id string, callback func(*body.VmRead) bool) {
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

func withVM(t *testing.T, requestBody body.VmCreate) body.VmRead {
	resp := e2e.DoPostRequest(t, "/vms", requestBody)
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "deployment was not created") {
		assert.FailNow(t, "deployment was not created")
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

	waitForJobFinished(t, vmCreated.JobID, func(jobRead *body.JobRead) bool {
		return true
	})

	waitForVmRunning(t, vmCreated.ID, func(vmRead *body.VmRead) bool {
		//make sure it is accessible
		if vmRead.ConnectionString != nil {
			return checkUpVM(t, *vmRead.ConnectionString)
		}
		return false
	})

	var vmRead body.VmRead
	readResp := e2e.DoGetRequest(t, "/vms/"+vmCreated.ID)
	err = e2e.ReadResponseBody(t, readResp, &vmRead)
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

func withAssumedFailedVM(t *testing.T, requestBody body.VmCreate) {
	resp := e2e.DoPostRequest(t, "/vms", requestBody)
	if resp.StatusCode == http.StatusBadRequest {
		return
	}

	var vmCreated body.VmCreated
	err := e2e.ReadResponseBody(t, resp, &vmCreated)
	assert.NoError(t, err, "created body was not read")

	t.Cleanup(func() {
		resp = e2e.DoDeleteRequest(t, "/vms/"+vmCreated.ID)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			assert.FailNow(t, "resource was not deleted")
		}

		waitForVmRunning(t, vmCreated.ID, func(vmRead *body.VmRead) bool {
			return true
		})
	})

	assert.FailNow(t, "resource was created but should have failed")
}
