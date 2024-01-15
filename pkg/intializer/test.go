package intializer

import (
	"fmt"
	"github.com/google/uuid"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/job"
	teamModels "go-deploy/models/sys/team"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service/user_service"
	"log"
	"time"
)

// CleanUpOldTests cleans up old E2E tests.
// Some E2E tests may fail and leave resources behind.
func CleanUpOldTests() {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Minute)

	oldE2eDeployments, err := deploymentModels.New().OlderThan(oneHourAgo).WithNameRegex("e2e-*").List()
	if err != nil {
		panic(fmt.Errorf("failed to list old e2e deployments: %w", err))
	}

	for _, deployment := range oldE2eDeployments {
		_ = job.New().Create(uuid.NewString(), "system", job.TypeDeleteDeployment, map[string]interface{}{
			"id": deployment.ID,
		})
	}

	oldE2eVms, err := vmModels.New().OlderThan(oneHourAgo).WithNameRegex("e2e-*").List()
	if err != nil {
		panic(fmt.Errorf("failed to list old e2e vms: %w", err))
	}

	for _, vm := range oldE2eVms {
		_ = job.New().Create(uuid.NewString(), "system", job.TypeDeleteVM, map[string]interface{}{
			"id": vm.ID,
		})

	}

	oldE2eTeams, err := teamModels.New().OlderThan(oneHourAgo).WithNameRegex("e2e-*").List()
	if err != nil {
		panic(fmt.Errorf("failed to list old e2e teams: %w", err))
	}

	for _, team := range oldE2eTeams {
		err := user_service.New().DeleteTeam(team.ID)
		if err != nil {
			panic(fmt.Errorf("failed to delete team %s: %w", team.ID, err))
		}
	}

	log.Println("e2e-tests cleanup:\n\t- deployments:", len(oldE2eDeployments), "\n\t- vms:", len(oldE2eVms), "\n\t- teams:", len(oldE2eTeams))
}
