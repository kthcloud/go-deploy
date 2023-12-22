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
	"time"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestList(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/vms")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var vms []body.VmRead
	err := e2e.ReadResponseBody(t, resp, &vms)
	assert.NoError(t, err, "vms were not fetched")

	for _, vm := range vms {
		assert.NotEmpty(t, vm.ID, "vm id was empty")
		assert.NotEmpty(t, vm.Name, "vm name was empty")
	}
}

func TestListGPUs(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/gpus")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gpus []body.GpuRead
	err := e2e.ReadResponseBody(t, resp, &gpus)
	assert.NoError(t, err, "gpus were not fetched")

	for _, gpu := range gpus {
		assert.NotEmpty(t, gpu.ID, "gpu id was empty")
		assert.NotEmpty(t, gpu.Name, "gpu name was empty")
	}
}

func TestListAvailableGPUs(t *testing.T) {
	t.Skip()
	return

	resp := e2e.DoGetRequest(t, "/gpus?available=true")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gpus []body.GpuRead
	err := e2e.ReadResponseBody(t, resp, &gpus)
	assert.NoError(t, err, "gpus were not fetched")

	for _, gpu := range gpus {
		assert.NotEmpty(t, gpu.ID, "gpu id was empty")
		assert.NotEmpty(t, gpu.Name, "gpu name was empty")

		available := gpu.Lease == nil || gpu.Lease.Expired
		assert.True(t, available, "gpu was not available")
	}
}

func TestCreate(t *testing.T) {
	publicKey := withSshPublicKey(t)

	requestBody := body.VmCreate{
		Name:         e2e.GenName("e2e"),
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

	_ = e2e.WithVM(t, requestBody)
}

func TestCreateWithInvalidBody(t *testing.T) {
	longName := body.VmCreate{
		Name:     "e2e-",
		CpuCores: 2,
		RAM:      2,
		DiskSize: 20,
	}
	for i := 0; i < 1000; i++ {
		longName.Name += uuid.NewString()
	}
	e2e.WithAssumedFailedVM(t, longName)

	invalidNames := []string{
		e2e.GenName("e2e") + "-",
		e2e.GenName("e2e") + "- ",
		e2e.GenName("e2e") + ".",
		"." + e2e.GenName("e2e"),
		e2e.GenName("e2e") + " " + e2e.GenName("e2e"),
		e2e.GenName("e2e") + "%",
		e2e.GenName("e2e") + "!",
		e2e.GenName("e2e") + "%" + e2e.GenName("e2e"),
	}

	for _, name := range invalidNames {
		requestBody := body.VmCreate{
			Name:     name,
			CpuCores: 2,
			RAM:      2,
			DiskSize: 20,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}

	invalidPorts := []body.Port{
		{
			Name:     strings.Repeat(uuid.NewString(), 100),
			Port:     100,
			Protocol: "tcp",
		},
		{
			Name:     "e2e-test",
			Port:     100,
			Protocol: "invalid",
		},
		{
			Name:     "e2e-test",
			Port:     -1,
			Protocol: "tcp",
		},
		{
			Name:     "e2e-test",
			Port:     100000,
			Protocol: "tcp",
		},
	}

	for _, port := range invalidPorts {
		requestBody := body.VmCreate{
			Name:         e2e.GenName("e2e"),
			SshPublicKey: withSshPublicKey(t),
			Ports: []body.Port{
				port,
			},
			CpuCores: 2,
			RAM:      2,
			DiskSize: 20,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}

	invalidCpuCores := []int{
		-1,
		0,
	}

	for _, cpuCores := range invalidCpuCores {
		requestBody := body.VmCreate{
			Name:         e2e.GenName("e2e"),
			SshPublicKey: withSshPublicKey(t),
			CpuCores:     cpuCores,
			RAM:          2,
			DiskSize:     20,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}

	invalidRam := []int{
		-1,
		0,
	}

	for _, ram := range invalidRam {
		requestBody := body.VmCreate{
			Name:         e2e.GenName("e2e"),
			SshPublicKey: withSshPublicKey(t),
			CpuCores:     2,
			RAM:          ram,
			DiskSize:     20,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}

	invalidDiskSize := []int{
		-1,
		0,
		10,
	}

	for _, diskSize := range invalidDiskSize {
		requestBody := body.VmCreate{
			Name:         e2e.GenName("e2e"),
			SshPublicKey: withSshPublicKey(t),
			CpuCores:     2,
			RAM:          2,
			DiskSize:     diskSize,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}

	invalidPublicKey := []string{
		"invalid",
		"ssh-rsa invalid",
		"ssh-rsa AAAAB3NzaC1yc2E AAAAB3NzaC1yc2E",
	}

	for _, publicKey := range invalidPublicKey {
		requestBody := body.VmCreate{
			Name:         e2e.GenName("e2e"),
			SshPublicKey: publicKey,
			CpuCores:     2,
			RAM:          2,
			DiskSize:     20,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}
}

func TestUpdate(t *testing.T) {
	publicKey := withSshPublicKey(t)

	requestBody := body.VmCreate{
		Name:         e2e.GenName("e2e"),
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

	vm := e2e.WithVM(t, requestBody)

	updatedPorts := []body.Port{
		{
			Name:     "e2e-test",
			Port:     100,
			Protocol: "tcp",
		},
		{
			Name:     "e2e-test-2",
			Port:     200,
			Protocol: "tcp",
		},
	}
	updatedCpuCores := 4
	updatedRam := 4

	updateRequestBody := body.VmUpdate{
		Ports:    &updatedPorts,
		CpuCores: &updatedCpuCores,
		RAM:      &updatedRam,
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID, updateRequestBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var vmUpdated body.VmUpdated
	err := e2e.ReadResponseBody(t, resp, &vmUpdated)
	assert.NoError(t, err, "vm was not updated")

	// make sure the job is picked up

	if vmUpdated.JobID != nil {
		e2e.WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	e2e.WaitForVmRunning(t, vm.ID, nil)

	var vmRead body.VmRead
	readResp := e2e.DoGetRequest(t, "/vms/"+vm.ID)
	err = e2e.ReadResponseBody(t, readResp, &vmRead)
	assert.NoError(t, err, "vm was not updated")

	if updateRequestBody.Ports != nil {
		for _, port := range *updateRequestBody.Ports {
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

	if updateRequestBody.CpuCores != nil {
		assert.Equal(t, updatedCpuCores, vmRead.Specs.CpuCores)
	}

	if updateRequestBody.RAM != nil {
		assert.Equal(t, updatedRam, vmRead.Specs.RAM)
	}
}

func TestCreateShared(t *testing.T) {
	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:      e2e.GenName("team"),
		Resources: []string{vm.ID},
		Members:   []body.TeamMemberCreate{{ID: e2e.PowerUserID}},
	})

	deploymentRead := e2e.GetDeployment(t, vm.ID)
	assert.Equal(t, []string{team.ID}, deploymentRead.Teams, "invalid teams on deployment")

	// Fetch team members deployments
	deployments := e2e.ListVMs(t, "?userId="+e2e.PowerUserID)
	assert.NotEmpty(t, deployments, "user has no deployments")

	hasDeployment := false
	for _, d := range deployments {
		if d.ID == vm.ID {
			hasDeployment = true
		}
	}

	assert.True(t, hasDeployment, "deployment was not found in other user's deployments")
}

func TestAttachAnyGPU(t *testing.T) {
	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	anyID := "any"

	updateGpuBody := body.VmUpdate{
		GpuID: &anyID,
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID, updateGpuBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var vmUpdated body.VmUpdated
	err := e2e.ReadResponseBody(t, resp, &vmUpdated)
	assert.NoError(t, err, "vm was not updated")

	// make sure the job is picked up
	time.Sleep(5 * time.Second)

	if vmUpdated.JobID != nil {
		e2e.WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	e2e.WaitForVmRunning(t, vm.ID, nil)

	var vmRead body.VmRead
	readResp := e2e.DoGetRequest(t, "/vms/"+vm.ID)
	err = e2e.ReadResponseBody(t, readResp, &vmRead)
	assert.NoError(t, err, "vm was not updated")

	// we can't check the gpu ID here, because it might be the case that
	// no gpus were actually available (reserved in another database)
}

func TestAttachGPU(t *testing.T) {
	// in order to test this, you need to set the gpu ID
	// this is done to prevent tests from "hogging" a single gpu
	// normally, it should be enough to just test with any gpu (as done above in TestAttachAnyGPU)
	gpuID := ""

	//goland:noinspection ALL
	if gpuID == "" {
		t.Skip("no gpu ID set")
	}

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	updateGpuBody := body.VmUpdate{
		GpuID: &gpuID,
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID, updateGpuBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var vmUpdated body.VmUpdated
	err := e2e.ReadResponseBody(t, resp, &vmUpdated)
	assert.NoError(t, err, "vm was not updated")

	// make sure the job is picked up
	time.Sleep(5 * time.Second)

	if vmUpdated.JobID != nil {
		e2e.WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	e2e.WaitForVmRunning(t, vm.ID, nil)

	var vmRead body.VmRead
	readResp := e2e.DoGetRequest(t, "/vms/"+vm.ID)
	err = e2e.ReadResponseBody(t, readResp, &vmRead)
	assert.NoError(t, err, "vm was not updated")
}

func TestAttachGPUWithInvalidID(t *testing.T) {
	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	invalidID := "invalid"

	updateGpuBody := body.VmUpdate{
		GpuID: &invalidID,
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID, updateGpuBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAttachGpuWithAlreadyAttachedID(t *testing.T) {
	// in order to test this, you need to set the gpu ID
	// this is done to prevent tests from "hogging" a single gpu
	// normally, it should be enough to just test with any gpu (as done above in TestAttachAnyGPU)
	gpuID := ""
	anotherGpuID := ""

	//goland:noinspection ALL
	if gpuID == "" {
		t.Skip("no gpu ID set")
	}

	//goland:noinspection ALL
	if anotherGpuID == "" {
		t.Skip("no another gpu ID set")
	}

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	updateGpuBody := body.VmUpdate{
		GpuID: &gpuID,
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID, updateGpuBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var vmUpdated body.VmUpdated
	err := e2e.ReadResponseBody(t, resp, &vmUpdated)
	assert.NoError(t, err, "vm was not updated")

	// make sure the job is picked up
	time.Sleep(5 * time.Second)

	if vmUpdated.JobID != nil {
		e2e.WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	e2e.WaitForVmRunning(t, vm.ID, nil)

	updateGpuBody = body.VmUpdate{
		GpuID: &anotherGpuID,
	}

	resp = e2e.DoPostRequest(t, "/vms/"+vm.ID, updateGpuBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCommand(t *testing.T) {
	commands := []string{"stop", "start", "reboot"}

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	for _, command := range commands {
		reqBody := body.VmCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/vms/"+vm.ID+"/command", reqBody)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		time.Sleep(30 * time.Second)
	}
}

func TestCreateAndRestoreSnapshot(t *testing.T) {
	publicKey := withSshPublicKey(t)

	requestBody := body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: publicKey,
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	}

	vm := e2e.WithVM(t, requestBody)

	snapshotCreateBody := body.VmSnapshotCreate{
		Name: e2e.GenName("e2e"),
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID+"/snapshots", snapshotCreateBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var snapshotCreated body.VmSnapshotCreated
	err := e2e.ReadResponseBody(t, resp, &snapshotCreated)
	assert.NoError(t, err, "snapshot was not created")
	assert.NotEmpty(t, snapshotCreated.ID)

	// TODO: add tests for snapshot delete

	// make sure the job is picked up
	time.Sleep(5 * time.Second)

	e2e.WaitForJobFinished(t, snapshotCreated.JobID, nil)
	e2e.WaitForVmRunning(t, vm.ID, nil)

	var vmSnapshotsRead []body.VmSnapshotRead
	readResp := e2e.DoGetRequest(t, "/vms/"+vm.ID+"/snapshots")
	err = e2e.ReadResponseBody(t, readResp, &vmSnapshotsRead)
	assert.NoError(t, err, "vm snapshots were not read")
	assert.NotEmpty(t, vmSnapshotsRead)

	var vmSnapshotRead body.VmSnapshotRead
	for _, snapshotRead := range vmSnapshotsRead {
		if snapshotRead.Name == snapshotCreateBody.Name {
			vmSnapshotRead = snapshotRead
			break
		}
	}

	assert.NotEmpty(t, vmSnapshotRead.ID)

	updateSnapshotBody := body.VmUpdate{
		SnapshotID: &vmSnapshotRead.ID,
	}

	resp = e2e.DoPostRequest(t, "/vms/"+vm.ID, updateSnapshotBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var vmUpdated body.VmUpdated
	err = e2e.ReadResponseBody(t, resp, &vmUpdated)
	assert.NoError(t, err, "vm was not updated")

	// make sure the job is picked up
	time.Sleep(5 * time.Second)

	if vmUpdated.JobID != nil {
		e2e.WaitForJobFinished(t, *vmUpdated.JobID, nil)
	}
	e2e.WaitForVmRunning(t, vm.ID, nil)

	var vmRead body.VmRead
	readResp = e2e.DoGetRequest(t, "/vms/"+vm.ID)
	err = e2e.ReadResponseBody(t, readResp, &vmRead)
	assert.NoError(t, err, "vm was not updated")
}

func TestInvalidCommand(t *testing.T) {
	invalidCommands := []string{"some command", "invalid"}

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName("e2e"),
		SshPublicKey: withSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	for _, command := range invalidCommands {
		reqBody := body.VmCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/vms/"+vm.ID+"/command", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}
