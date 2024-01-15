package vm

const (
	// ActivityBeingCreated is the activity used for VMs that are being created.
	ActivityBeingCreated = "beingCreated"
	// ActivityBeingDeleted is the activity used for VMs that are being deleted.
	ActivityBeingDeleted = "beingDeleted"
	// ActivityUpdating is the activity used for VMs that are being updated.
	ActivityUpdating = "updating"
	// ActivityAttachingGPU is the activity used for VMs that are attaching a GPU.
	ActivityAttachingGPU = "attachingGpu"
	// ActivityDetachingGPU is the activity used for VMs that are detaching a GPU.
	ActivityDetachingGPU = "detachingGpu"
	// ActivityRepairing is the activity used for VMs that are being repaired.
	ActivityRepairing = "repairing"
)
