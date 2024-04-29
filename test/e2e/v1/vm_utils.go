package v1

import (
	"fmt"
	"github.com/helloyi/go-sshclient"
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/test/e2e"
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	VmPath  = "/v1/vms/"
	VmsPath = "/v1/vms"
)

func GetVM(t *testing.T, id string, userID ...string) body.VmRead {
	resp := e2e.DoGetRequest(t, VmPath+id, userID...)
	return e2e.MustParse[body.VmRead](t, resp)
}

func ListVMs(t *testing.T, query string, userID ...string) []body.VmRead {
	resp := e2e.DoGetRequest(t, VmsPath+query, userID...)
	return e2e.MustParse[[]body.VmRead](t, resp)
}

func UpdateVM(t *testing.T, id string, requestBody body.VmUpdate, userID ...string) body.VmRead {
	resp := e2e.DoPostRequest(t, VmPath+id, requestBody, userID...)
	vmUpdated := e2e.MustParse[body.VmUpdated](t, resp)

	if vmUpdated.JobID != nil {
		WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	WaitForVmRunning(t, id, func(vmRead *body.VmRead) bool {
		// Make sure it is accessible
		if vmRead.ConnectionString != nil {
			return checkUpVM(t, *vmRead.ConnectionString)
		}
		return false
	})

	return GetVM(t, id, userID...)
}

func GetSnapshot(t *testing.T, vmID string, snapshotID string, userID ...string) body.VmSnapshotRead {
	resp := e2e.DoGetRequest(t, "/vms/"+vmID+"/snapshots/"+snapshotID, userID...)
	return e2e.MustParse[body.VmSnapshotRead](t, resp)
}

func ListSnapshots(t *testing.T, vmID string, userID ...string) []body.VmSnapshotRead {
	resp := e2e.DoGetRequest(t, "/vms/"+vmID+"/snapshots", userID...)
	return e2e.MustParse[[]body.VmSnapshotRead](t, resp)
}

func CreateSnapshot(t *testing.T, id string, requestBody body.VmSnapshotCreate, userID ...string) body.VmSnapshotRead {
	resp := e2e.DoPostRequest(t, VmPath+id+"/snapshots", requestBody, userID...)
	snapshotCreated := e2e.MustParse[body.VmSnapshotCreated](t, resp)

	WaitForJobFinished(t, snapshotCreated.JobID, nil)
	WaitForVmRunning(t, id, nil)

	snapshots := ListSnapshots(t, id, userID...)
	for _, snapshot := range snapshots {
		if snapshot.Name == requestBody.Name {
			return snapshot
		}
	}

	assert.Fail(t, "snapshot not found")
	return body.VmSnapshotRead{}
}

func DeleteSnapshot(t *testing.T, vmID string, snapshotID string, userID ...string) {
	resp := e2e.DoDeleteRequest(t, "/vms/"+vmID+"/snapshots/"+snapshotID, userID...)
	if resp.StatusCode == http.StatusNotFound {
		return
	}

	snapshotDeleted := e2e.MustParse[body.VmSnapshotDeleted](t, resp)
	WaitForJobFinished(t, snapshotDeleted.JobID, nil)
	WaitForVmRunning(t, vmID, nil)
}

func GetGPU(t *testing.T, gpuID string, userID ...string) body.GpuRead {
	resp := e2e.DoGetRequest(t, "/gpus/"+gpuID, userID...)
	return e2e.MustParse[body.GpuRead](t, resp)
}

func ListGPUs(t *testing.T, query string, userID ...string) []body.GpuRead {
	resp := e2e.DoGetRequest(t, "/gpus"+query, userID...)
	return e2e.MustParse[[]body.GpuRead](t, resp)
}

func WithDefaultVM(t *testing.T, userID ...string) body.VmRead {
	return WithVM(t, body.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: WithSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})
}

func WithVM(t *testing.T, requestBody body.VmCreate, userID ...string) body.VmRead {
	resp := e2e.DoPostRequest(t, VmsPath, requestBody, userID...)
	vmCreated := e2e.MustParse[body.VmCreated](t, resp)

	t.Cleanup(func() { cleanUpVm(t, vmCreated.ID) })

	WaitForJobFinished(t, vmCreated.JobID, nil)
	WaitForVmRunning(t, vmCreated.ID, func(vmRead *body.VmRead) bool {
		//make sure it is accessible
		if vmRead.ConnectionString != nil {
			return checkUpVM(t, *vmRead.ConnectionString)
		}
		return false
	})

	readResp := e2e.DoGetRequest(t, "/vms/"+vmCreated.ID, userID...)
	vmRead := e2e.MustParse[body.VmRead](t, readResp)

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

	assert.FailNow(t, "model was created but should have failed")
}

func WithSshPublicKey(t *testing.T) string {
	content, err := os.ReadFile("../../ssh/id_rsa.pub")
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

	client, err := sshclient.DialWithKey(fmt.Sprintf("%s:%s", address, port), user, "../../ssh/id_rsa")
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
	resp := e2e.DoDeleteRequest(t, VmPath+id)
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
