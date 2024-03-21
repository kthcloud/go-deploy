package vms

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/dto/v1/body"
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
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	queries := []string{
		"?page=1&pageSize=10",
		"?userId=" + e2e.PowerUserID + "&page=1&pageSize=3",
		"?userId=" + e2e.DefaultUserID + "&page=1&pageSize=3",
	}

	for _, query := range queries {
		e2e.ListVMs(t, query)
	}
}

func TestListGPUs(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	queries := []string{
		"?page=1&pageSize=3",
		"?available=true&page=1&pageSize=3",
	}

	for _, query := range queries {
		e2e.ListGPUs(t, query)
	}
}

func TestCreate(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	requestBody := body.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: e2e.WithSshPublicKey(t),
		Ports: []body.PortCreate{
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

	e2e.WithVM(t, requestBody)
}

func TestCreateWithInvalidBody(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

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
		e2e.GenName() + "-",
		e2e.GenName() + "- ",
		e2e.GenName() + ".",
		"." + e2e.GenName(),
		e2e.GenName() + " " + e2e.GenName(),
		e2e.GenName() + "%",
		e2e.GenName() + "!",
		e2e.GenName() + "%" + e2e.GenName(),
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

	invalidPorts := []body.PortCreate{
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
			Name:         e2e.GenName(),
			SshPublicKey: e2e.WithSshPublicKey(t),
			Ports: []body.PortCreate{
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
			Name:         e2e.GenName(),
			SshPublicKey: e2e.WithSshPublicKey(t),
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
			Name:         e2e.GenName(),
			SshPublicKey: e2e.WithSshPublicKey(t),
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
			Name:         e2e.GenName(),
			SshPublicKey: e2e.WithSshPublicKey(t),
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
			Name:         e2e.GenName(),
			SshPublicKey: publicKey,
			CpuCores:     2,
			RAM:          2,
			DiskSize:     20,
		}
		e2e.WithAssumedFailedVM(t, requestBody)
	}
}

func TestUpdate(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	vm := e2e.WithDefaultVM(t)

	updatedCpuCores := 4
	updatedRam := 4
	updateRequestBody := body.VmUpdate{
		Ports: &[]body.PortUpdate{
			{
				Name:     "e2e-test",
				Port:     100,
				Protocol: "tcp",
			},
		},
		CpuCores: &updatedCpuCores,
		RAM:      &updatedRam,
	}

	vm = e2e.UpdateVM(t, vm.ID, updateRequestBody)

	if updateRequestBody.Ports != nil {
		for _, port := range *updateRequestBody.Ports {
			found := false
			for _, portRead := range vm.Ports {
				if port.Name == portRead.Name {
					assert.Equal(t, port.Port, portRead.Port)
					assert.Equal(t, port.Protocol, portRead.Protocol)
					found = true
					break
				}
			}
			assert.True(t, found, "port not found")
		}
	}

	if updateRequestBody.CpuCores != nil {
		assert.Equal(t, updatedCpuCores, vm.Specs.CpuCores)
	}

	if updateRequestBody.RAM != nil {
		assert.Equal(t, updatedRam, vm.Specs.RAM)
	}

	// TODO: Make sure the VM actually has the new specs (e.g. by running a command over SSH)
}

func TestCreateShared(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	vm := e2e.WithDefaultVM(t)
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:      e2e.GenName(),
		Resources: []string{vm.ID},
		Members:   []body.TeamMemberCreate{{ID: e2e.PowerUserID}},
	})

	vmRead := e2e.GetVM(t, vm.ID)
	assert.Equal(t, []string{team.ID}, vmRead.Teams, "invalid teams on vm")

	// Fetch team members vms
	vms := e2e.ListVMs(t, "?userId="+e2e.PowerUserID)
	assert.NotEmpty(t, vms, "user has no vms")

	hasVM := false
	for _, d := range vms {
		if d.ID == vm.ID {
			hasVM = true
		}
	}

	assert.True(t, hasVM, "vm was not found in other user's vms")
}

func TestAttachAnyGPU(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	// TODO: Fix this test
	// This test is currently unreliable, as it might attach to a GPU that is leased in production
	//
	// We would need to dedicate a small GPU for testing, but this is not possible at the moment, since we have so few
	// GPUs available
	t.Skip("not enough GPUs available to run this test")
	return

	t.Parallel()

	vm := e2e.WithDefaultVM(t)

	anyID := "any"

	updateGpuBody := body.VmUpdate{
		GpuID: &anyID,
	}

	e2e.UpdateVM(t, vm.ID, updateGpuBody)

	// We can't check the GPU ID here, because it might be the case that
	// no GPUs were available (reserved in another database)

	// TODO: check that the GPU is actually attached by running a command over SSH (e.g. nvidia-smi or lspci)
}

func TestAttachGPU(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	// To test this, you need to set the gpu_repo ID
	// This is done to prevent tests from "hogging" a single gpu_repo
	// Normally, it should be enough to just test with any GPU (as done above in TestAttachAnyGPU)
	gpuID := ""

	//goland:noinspection ALL
	if gpuID == "" {
		t.Skip("no gpu_repo ID set")
	}

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: e2e.WithSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	updateGpuBody := body.VmUpdate{
		GpuID: &gpuID,
	}

	e2e.UpdateVM(t, vm.ID, updateGpuBody)

	// We can't check the GPU ID here, because it might be the case that
	// the GPU was not available (reserved in another database)

	// TODO: check that the GPU is actually attached by running a command over SSH (e.g. nvidia-smi or lspci)
}

func TestAttachGPUWithInvalidID(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: e2e.WithSshPublicKey(t),
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
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	// To test this, you need to set the gpu_repo ID
	// This is done to prevent tests from "hogging" a single gpu_repo
	// Normally, it should be enough to just test with any gpu_repo (as done above in TestAttachAnyGPU)
	gpuID := ""
	anotherGpuID := ""

	//goland:noinspection ALL
	if gpuID == "" {
		t.Skip("no gpu_repo ID set")
	}

	//goland:noinspection ALL
	if anotherGpuID == "" {
		t.Skip("no another gpu_repo ID set")
	}

	vm := e2e.WithVM(t, body.VmCreate{
		Name:         e2e.GenName(),
		SshPublicKey: e2e.WithSshPublicKey(t),
		CpuCores:     2,
		RAM:          2,
		DiskSize:     20,
	})

	updateGpuBody := body.VmUpdate{
		GpuID: &gpuID,
	}

	e2e.UpdateVM(t, vm.ID, updateGpuBody)

	updateGpuBody = body.VmUpdate{
		GpuID: &anotherGpuID,
	}

	resp := e2e.DoPostRequest(t, "/vms/"+vm.ID, updateGpuBody)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCommand(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	commands := []string{"stop", "start", "reboot"}
	vm := e2e.WithDefaultVM(t)

	for _, command := range commands {
		reqBody := body.VmCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/vms/"+vm.ID+"/command", reqBody)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		time.Sleep(30 * time.Second)
	}
}

func TestCreateAndRestoreSnapshot(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	vm := e2e.WithDefaultVM(t)

	snapshotCreateBody := body.VmSnapshotCreate{
		Name: e2e.GenName(),
	}

	// Create
	snapshot := e2e.CreateSnapshot(t, vm.ID, snapshotCreateBody)
	t.Cleanup(func() {
		e2e.DeleteSnapshot(t, vm.ID, snapshot.ID)
	})

	// Get
	snapshot = e2e.GetSnapshot(t, vm.ID, snapshot.ID)

	// List
	snapshots := e2e.ListSnapshots(t, vm.ID)
	assert.NotEmpty(t, snapshots, "no snapshots found")
	found := false
	for _, s := range snapshots {
		if snapshot.ID == s.ID {
			assert.Equal(t, snapshotCreateBody.Name, s.Name)
			found = true
			break
		}
	}
	assert.True(t, found, "snapshot not found in list")

	// Restore
	// Edit the VM to make sure it can be restored
	e2e.DoSshCommand(t, *vm.ConnectionString, "echo 'e2e-test' > /tmp/test.txt")

	restoreSnapshotBody := body.VmUpdate{
		SnapshotID: &snapshot.ID,
	}

	e2e.UpdateVM(t, vm.ID, restoreSnapshotBody)

	// Check that the file is not there
	res := e2e.DoSshCommand(t, *vm.ConnectionString, "cat /tmp/test.txt")
	if strings.Contains(res, "e2e-test") {
		assert.Fail(t, "snapshot did not restore correctly")
	}
}

func TestInvalidCommand(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	invalidCommands := []string{"some command", "invalid"}

	vm := e2e.WithDefaultVM(t)

	for _, command := range invalidCommands {
		reqBody := body.VmCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/vms/"+vm.ID+"/command", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}
