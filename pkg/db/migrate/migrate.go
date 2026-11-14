//lint:file-ignore // Ignore (SA1019) since we want to migrate deprecated.
package migrator

import (
	"context"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// Migrate runs every migration specified in the getMigrations function.
// It should be run as early as possible in the program, and should never be called again.
func Migrate(ctx context.Context, logger *zap.Logger) error {
	migrations := getMigrations()

	if logger == nil {
		logger = zap.NewNop()
	}
	slogger := logger.Sugar()

	if len(migrations) > 0 {
		for name, migration := range migrations {
			slogger.Infof("- %s (%d/%d)", name, 1, len(migrations))
			if err := migration(ctx, logger.Named(name)); err != nil {
				return fmt.Errorf("migration %s failed. details: %w", name, err)
			}
		}
	} else {
		slogger.Infoln("No migrations to run")
	}

	return nil
}

// getMigrations returns a map of migrations to run.
// add a migration to the list of functions to run.
// clear when prod has run it once.
//
// the migrations must be **idempotent**.
//
// add a date to the migration name to make it easier to identify.
func getMigrations() map[string]func(ctx context.Context, logger *zap.Logger) error {
	return map[string]func(ctx context.Context, logger *zap.Logger) error{
		"migratePrivateBooleanToVisibilityEnum_2024_06_10": migratePrivateBooleanToVisibilityEnum_2024_06_10,
	}
}

func migratePrivateBooleanToVisibilityEnum_2024_06_10(ctx context.Context, logger *zap.Logger) error {
	var start = time.Now()
	deployments, err := deployment_repo.New().List()
	if err != nil {
		return err
	}

	migrated := len(deployments)

	for _, deployment := range deployments {
		mainApp := deployment.GetMainApp()

		if mainApp.Visibility != "" {
			migrated--
			continue
		}

		if mainApp.Private { //nolint:staticcheck // SA1019: migrating from deprecated Private field
			mainApp.Visibility = model.VisibilityPrivate
		} else {
			mainApp.Visibility = model.VisibilityPublic
		}

		err = deployment_repo.New().SetWithBsonByID(deployment.ID, bson.D{{Key: "apps.main.visibility", Value: mainApp.Visibility}})
		if err != nil {
			return err
		}
	}

	logger.Info("migrated", zap.Int("deploymentCount", migrated), zap.Duration("in", time.Since(start)))

	return nil
}
