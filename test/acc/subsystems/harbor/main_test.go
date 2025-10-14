package harbor

import (
	"github.com/kthcloud/go-deploy/test/acc"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	acc.Setup()
	if err := os.Chdir("../../../.."); err != nil {
		log.Fatalf("Failed to change working directory: %v", err)
	}
	code := m.Run()
	acc.Shutdown()
	os.Exit(code)
}
