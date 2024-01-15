package deployment

const (
	// ActivityBeingCreated is used when a deployment is being created.
	ActivityBeingCreated = "beingCreated"

	// ActivityBeingDeleted is used when a deployment is being deleted.
	ActivityBeingDeleted = "beingDeleted"

	// ActivityUpdating is used when a deployment is being updated.
	ActivityUpdating = "updating"

	// ActivityRestarting is used when a deployment is being restarted.
	ActivityRestarting = "restarting"

	// ActivityBuilding is used when a deployment is being built.
	ActivityBuilding = "building"

	// ActivityRepairing is used when a deployment is being repaired.
	ActivityRepairing = "repairing"
)
