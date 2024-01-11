package cs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"testing"
)

func TestReadHost(t *testing.T) {
	// This assumes the host exists
	// This will obviously fail if the host is renamed or deleted
	hostName := "se-flem-001"

	client := withClient(t)
	host, err := client.ReadHostByName(hostName)
	test.NoError(t, err, "failed to read host")

	assert.Equal(t, hostName, host.Name, "host name does not match")
}

func TestReadVmHostname(t *testing.T) {
	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t)

	hostname, err := client.ReadHostByVM(vm.ID)
	test.NoError(t, err, "failed to read vm hostname")

	assert.NotEmpty(t, hostname, "vm hostname is empty")
}
