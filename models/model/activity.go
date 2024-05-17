package model

import "time"

type Activity struct {
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdAt"`
}

const (
	// ActivityBeingCreated is used when a model is being created.
	ActivityBeingCreated = "beingCreated"

	// ActivityBeingDeleted is used when a model is being deleted.
	ActivityBeingDeleted = "beingDeleted"

	// ActivityUpdating is used when a model is being updated.
	ActivityUpdating = "updating"

	// ActivityRestarting is used when a model is being restarted.
	ActivityRestarting = "restarting"

	// ActivityBuilding is used when a model is being built.
	ActivityBuilding = "building"

	// ActivityRepairing is used when a model is being repaired.
	ActivityRepairing = "repairing"
)
