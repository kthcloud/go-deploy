package vms

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func withSshPublicKey(t *testing.T) string {
	content, err := os.ReadFile("../../ssh/id_rsa.pub")
	assert.NoError(t, err, "could not read ssh public key")
	return strings.TrimSpace(string(content))
}
