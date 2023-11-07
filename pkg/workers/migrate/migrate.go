package migrator

import (
	"context"
	"go-deploy/models/db"
	"go-deploy/models/sys/activity"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

// Migrate will run as early as possible in the program, and it will never be called again.
func Migrate() {
	migrations := getMigrations()

	if len(migrations) > 0 {
		log.Println("migrating...")

		for name, migration := range migrations {
			log.Printf("running migration %s", name)
			if err := migration(); err != nil {
				log.Fatalf("failed to run migration %s. details: %s", name, err)
			}
		}

		log.Println("migrations done")
		return
	}

	log.Println("nothing to migrate")
}

// getMigrations returns a map of migrations to run.
// add a migration to the list of functions to run.
// clear when prod has run it once.
//
// the migrations must be **idempotent**.
//
// add a date to the migration name to make it easier to identify.
func getMigrations() map[string]func() error {
	return map[string]func() error{
		"migration_2021_10_07": migrateActivitesToNewModel_2023_11_07,
	}
}

type withOldActivites struct {
	ID         string   `bson:"id"`
	Activities []string `bson:"activities"`
}

func migrateActivitesToNewModel_2023_11_07() error {

	cursor, err := db.DB.GetCollection("deployments").Find(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	for cursor.Next(context.Background()) {
		var deployment withOldActivites
		if err := cursor.Decode(&deployment); err != nil {
			// skip this in case we already migrated it
		}

		activities := make(map[string]activity.Activity)
		for _, a := range deployment.Activities {
			activities[a] = activity.Activity{
				Name:      a,
				CreatedAt: time.Now(),
			}
		}

		_, err = db.DB.GetCollection("deployments").UpdateOne(context.Background(), bson.M{"id": deployment.ID}, bson.M{"$set": bson.M{"activities": activities}})
		if err != nil {
			return err
		}
	}

	cursor, err = db.DB.GetCollection("vms").Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	for cursor.Next(context.Background()) {
		var vm withOldActivites
		if err := cursor.Decode(&vm); err != nil {
			// skip this in case we already migrated it
		}

		activities := make(map[string]activity.Activity)
		for _, a := range vm.Activities {
			activities[a] = activity.Activity{
				Name:      a,
				CreatedAt: time.Now(),
			}
		}

		_, err = db.DB.GetCollection("vms").UpdateOne(context.Background(), bson.M{"id": vm.ID}, bson.M{"$set": bson.M{"activities": activities}})
		if err != nil {
			return err
		}
	}

	cursor, err = db.DB.GetCollection("storageManagers").Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	for cursor.Next(context.Background()) {
		var storageManager withOldActivites
		if err := cursor.Decode(&storageManager); err != nil {
			// skip this in case we already migrated it
		}

		activities := make(map[string]activity.Activity)
		for _, a := range storageManager.Activities {
			activities[a] = activity.Activity{
				Name:      a,
				CreatedAt: time.Now(),
			}
		}

		_, err = db.DB.GetCollection("storageManagers").UpdateOne(context.Background(), bson.M{"id": storageManager.ID}, bson.M{"$set": bson.M{"activities": activities}})
		if err != nil {
			return err
		}
	}

	return nil
}
