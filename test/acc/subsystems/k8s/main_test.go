package k8s

import (
	"go-deploy/test/acc"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	acc.Setup()
	code := m.Run()
	acc.Shutdown()
	os.Exit(code)
}
