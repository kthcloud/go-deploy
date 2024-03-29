package intializer

import (
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/models/versions"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/db/resources/team_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/service"
	"log"
	"time"
)

// CleanUpOldTests cleans up old E2E tests.
// Some E2E tests may fail and leave resources behind.
func CleanUpOldTests() {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Minute)

	oldE2eDeployments, err := deployment_repo.New().OlderThan(oneHourAgo).WithNameRegex("e2e-*").List()
	if err != nil {
		panic(fmt.Errorf("failed to list old e2e deployments: %w", err))
	}

	for _, deployment := range oldE2eDeployments {
		_ = job_repo.New().Create(uuid.NewString(), "system", model.JobDeleteDeployment, versions.V1, map[string]interface{}{
			"id": deployment.ID,
		})
	}

	oldE2eVms, err := vm_repo.New().OlderThan(oneHourAgo).WithNameRegex("e2e-*").List()
	if err != nil {
		panic(fmt.Errorf("failed to list old e2e vms: %w", err))
	}

	for _, vm := range oldE2eVms {
		_ = job_repo.New().Create(uuid.NewString(), "system", model.JobDeleteVM, versions.V1, map[string]interface{}{
			"id": vm.ID,
		})

	}

	oldE2eTeams, err := team_repo.New().OlderThan(oneHourAgo).WithNameRegex("e2e-*").List()
	if err != nil {
		panic(fmt.Errorf("failed to list old e2e teams: %w", err))
	}

	for _, team := range oldE2eTeams {
		err := service.V1().Teams().Delete(team.ID)
		if err != nil {
			panic(fmt.Errorf("failed to delete team %s: %w", team.ID, err))
		}
	}

	log.Println("e2e-tests cleanup:\n\t- deployments:", len(oldE2eDeployments), "\n\t- vms:", len(oldE2eVms), "\n\t- teams:", len(oldE2eTeams))
}
