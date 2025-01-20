package v2

import (
	"fmt"
	"github.com/helloyi/go-sshclient"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	VmPath  = "/v2/vms/"
	VmsPath = "/v2/vms"

	VmActionsPath = "/v2/vmActions"
)

func GetVM(t *testing.T, id string, user ...string) body.VmRead {
	resp := e2e.DoGetRequest(t, VmPath+id, user...)
	return e2e.MustParse[body.VmRead](t, resp)
}

func ListVMs(t *testing.T, query string, user ...string) []body.VmRead {
	resp := e2e.DoGetRequest(t, VmsPath+query, user...)
	return e2e.MustParse[[]body.VmRead](t, resp)
}

func UpdateVM(t *testing.T, id string, requestBody body.VmUpdate, user ...string) body.VmRead {
	resp := e2e.DoPostRequest(t, VmPath+id, requestBody, user...)
	vmUpdated := e2e.MustParse[body.VmUpdated](t, resp)

	if vmUpdated.JobID != nil {
		WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	WaitForVmRunning(t, id, func(vmRead *body.VmRead) bool {
		// Make sure it is accessible
		if vmRead.SshConnectionString != nil {
			return checkUpVM(t, *vmRead.SshConnectionString)
		}
		return false
	})

	return GetVM(t, id, user...)
}

func DoVmAction(t *testing.T, vmID string, requestBody body.VmActionCreate, user ...string) body.VmActionCreated {
	resp := e2e.DoPostRequest(t, VmActionsPath+"?vmId="+vmID, requestBody, user...)
	vmActionCreated := e2e.MustParse[body.VmActionCreated](t, resp)

	WaitForJobFinished(t, vmActionCreated.JobID, nil)

	return vmActionCreated
}

func WithDefaultVM(t *testing.T, userID ...string) body.VmRead {
	return WithVM(t, body.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: WithSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     10,
	}, userID...)
}

func WithVM(t *testing.T, requestBody body.VmCreate, user ...string) body.VmRead {
	resp := e2e.DoPostRequest(t, VmsPath, requestBody, user...)
	vmCreated := e2e.MustParse[body.VmCreated](t, resp)

	t.Cleanup(func() { cleanUpVm(t, vmCreated.ID) })

	WaitForJobFinished(t, vmCreated.JobID, nil)
	WaitForVmRunning(t, vmCreated.ID, func(vmRead *body.VmRead) bool {
		// Make sure it is accessible
		if vmRead.SshConnectionString != nil {
			return checkUpVM(t, *vmRead.SshConnectionString)
		}
		return false
	})

	vmRead := GetVM(t, vmCreated.ID, user...)

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
	resp := e2e.DoPostRequest(t, VmsPath, requestBody)
	if resp.StatusCode == http.StatusBadRequest {
		return
	}

	var vmCreated body.VmCreated
	err := e2e.ReadResponseBody(t, resp, &vmCreated)
	assert.NoError(t, err, "created body was not read")

	t.Cleanup(func() { cleanUpVm(t, vmCreated.ID) })

	assert.FailNow(t, "vm was created but should have failed")
}

func WithSshPublicKey(t *testing.T) string {
	content, err := os.ReadFile("../../../ssh/id_rsa.pub")
	assert.NoError(t, err, "could not read ssh public key")
	return strings.TrimSpace(string(content))
}

func DoSshCommand(t *testing.T, connectionString, cmd string) string {
	// ssh user@address -p port
	connectionStringParts := strings.Split(connectionString, " ")
	assert.Len(t, connectionStringParts, 4)

	addrParts := strings.Split(connectionStringParts[1], "@")
	assert.Len(t, addrParts, 2)

	user := addrParts[0]
	address := addrParts[1]
	port := connectionStringParts[3]

	client, err := sshclient.DialWithKey(fmt.Sprintf("%s:%s", address, port), user, "../../../ssh/id_rsa")
	if err != nil || client == nil {
		return ""
	}
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	output, _ := client.Cmd(cmd).SmartOutput()
	return string(output)
}

func WaitForVmRunning(t *testing.T, id string, callback func(*body.VmRead) bool) {
	e2e.FetchUntil(t, VmPath+id, func(resp *http.Response) bool {
		vmRead := e2e.MustParse[body.VmRead](t, resp)
		if vmRead.Status == status_codes.GetMsg(status_codes.ResourceRunning) {
			if callback == nil || callback(&vmRead) {
				return true
			}
		}

		return false
	})
}

func WaitForVmDeleted(t *testing.T, id string, callback func() bool) {
	e2e.FetchUntil(t, VmPath+id, func(resp *http.Response) bool {
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

	output := DoSshCommand(t, connectionString, "echo 'hello world'")
	return output == "hello world\n"
}

func cleanUpVm(t *testing.T, id string) {
	resp := e2e.DoDeleteRequest(t, VmPath+id, e2e.AdminUser)
	if resp.StatusCode == http.StatusNotFound {
		return
	}

	if resp.StatusCode == http.StatusOK {
		var vmDeleted body.VmDeleted
		err := e2e.ReadResponseBody(t, resp, &vmDeleted)
		assert.NoError(t, err, "deleted body was not read")
		assert.Equal(t, id, vmDeleted.ID)

		WaitForJobFinished(t, vmDeleted.JobID, nil)
		WaitForVmDeleted(t, vmDeleted.ID, nil)

		return
	}

	assert.FailNow(t, "vm was not deleted")
}
