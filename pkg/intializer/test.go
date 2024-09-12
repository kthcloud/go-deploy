package intializer

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/models/mode"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	rErrors "github.com/kthcloud/go-deploy/pkg/db/resources/errors"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/team_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/user_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/service"
	"time"
)

// CleanUpOldTests cleans up old E2E tests.
// Some E2E tests may fail and leave resources behind.
func CleanUpOldTests() error {
	now := time.Now()
	oldTestThreshold := now.Add(-1 * time.Second)

	oldE2eDeployments, err := deployment_repo.New().OlderThan(oldTestThreshold).WithNameRegex("e2e-*").List()
	if err != nil {
		return fmt.Errorf("failed to list old e2e deployments: %w", err)
	}

	for _, deployment := range oldE2eDeployments {
		_ = job_repo.New().Create(uuid.NewString(), "system", model.JobDeleteDeployment, version.V2, map[string]interface{}{
			"id": deployment.ID,
		})
	}

	oldE2eVms, err := vm_repo.New().OlderThan(oldTestThreshold).WithNameRegex("e2e-*").List()
	if err != nil {
		return fmt.Errorf("failed to list old e2e vms: %w", err)
	}

	for _, vm := range oldE2eVms {
		if vm.Version == version.V2 {
			_ = job_repo.New().Create(uuid.NewString(), "system", model.JobDeleteVM, version.V2, map[string]interface{}{
				"id": vm.ID,
			})
		}
	}

	oldE2eTeams, err := team_repo.New().OlderThan(oldTestThreshold).WithNameRegex("e2e-*").List()
	if err != nil {
		return fmt.Errorf("failed to list old e2e teams: %w", err)
	}

	for _, team := range oldE2eTeams {
		err := service.V2().Teams().Delete(team.ID)
		if err != nil {
			return fmt.Errorf("failed to delete team %s: %w", team.ID, err)
		}
	}

	return nil
}

// EnsureTestUsersExist ensures that the test users are created.
func EnsureTestUsersExist() error {
	if config.Config.Mode != mode.Test {
		return nil
	}

	users, err := service.V2().Users().ListTestUsers()
	if err != nil {
		return fmt.Errorf("failed to list test users: %w", err)
	}

	for _, user := range users {
		_, err = user_repo.New().Synchronize(user.ID, &model.UserSynchronizeParams{
			Username:      user.Username,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Email:         user.Email,
			IsAdmin:       user.IsAdmin,
			EffectiveRole: &user.EffectiveRole,
		})
		if err != nil && !errors.Is(err, rErrors.NonUniqueFieldErr) {
			return fmt.Errorf("failed to synchronize user %s: %w", user.ID, err)
		}

		// Ensure test user's API key matches
		err = user_repo.New().UpdateWithParams(user.ID, &model.UserUpdateParams{
			ApiKeys: &user.ApiKeys,
		})
		if err != nil {
			return fmt.Errorf("failed to update user %s: %w", user.ID, err)
		}

		u, err := user_repo.New().GetByID(user.ID)
		if err != nil {
			return fmt.Errorf("failed to get user %s: %w", user.ID, err)
		}

		log.Printf("Added test user %s (API-key: %s)", u.Username, u.ApiKeys[0].Key)
	}

	return nil
}
