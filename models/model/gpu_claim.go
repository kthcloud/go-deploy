package model

// GpuClaim represents a claim for gpus that deployments can use
type GpuClaim struct {
	Name      string             `bson:"name"`
	Zone      string             `bson:"zone"`
	Requested []RequestedGpu     `bson:"requested"`
	Allocated []AllocatedGpu     `bson:"allocated,omitempty"`
	Consumers []GpuClaimConsumer `bson:"consumers,omitempty"`
}

type RequestAllocationMode string

type RequestedGpu struct {
	AllocationMode  RequestAllocationMode `bson:"allocationMode"`
	Capacity        map[string]string     `bson:"capacity,omitempty"`
	Count           *int64                `bson:"count,omitempty"`
	DeviceClassName string                `bson:"deviceClassName"`
	Selectors       []string              `bson:"selectors,omitempty"`
}

type AllocatedGpu struct {
	Pool        string `bson:"pool,omitempty"`
	Device      string `bson:"device,omitempty"`
	Request     string `bson:"request,omitempty"`
	ShareID     string `bson:"shareID,omitempty"`
	AdminAccess bool   `bson:"adminAccess,omitempty"`
}

type GpuClaimConsumer struct {
	APIGroup string `bson:"apiGroup,omitempty"`
	Resource string `bson:"resource,omitempty"`
	Name     string `bson:"name,omitempty"`
	UID      string `bson:"uid,omitempty"`
}
