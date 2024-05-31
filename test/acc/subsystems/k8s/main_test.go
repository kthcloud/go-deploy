package k8s

import (
	"go-deploy/test/acc"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	acc.Setup()
	code := m.Run()
	if err := os.Chdir("../../../.."); err != nil {
		log.Fatalf("Failed to change working directory: %v", err)
	}
	acc.Shutdown()
	os.Exit(code)
}
