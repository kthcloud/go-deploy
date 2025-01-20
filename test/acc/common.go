package acc

import (
	"fmt"
	"github.com/kthcloud/go-deploy/models/mode"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/log"
	"os"
	"strings"

	"github.com/google/uuid"
)

const (
	VmTestsEnabled = false
)

func Setup() {
	err := log.SetupLogger(mode.Test)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup logger. details: %s", err.Error()))
	}

	if err := os.Chdir("../../../.."); err != nil {
		log.Fatalf("Failed to change working directory: %v", err)
	}

	err = config.SetupEnvironment(mode.Test)
	if err != nil {
		log.Fatalln(err)
	}

	config.Config.MongoDB.Name = config.Config.MongoDB.Name + "-test"
}

func Shutdown() {

}

func GenName(base ...string) string {
	if len(base) == 0 {
		return "acc-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
	}

	return "acc-" + strings.ReplaceAll(base[0], " ", "-") + "-" + strings.ReplaceAll(uuid.NewString()[:10], "-", "")
}
