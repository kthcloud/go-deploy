package model

type GpuClaimCreateParams struct {
	Name string
	Zone string

	Requested map[string]RequestedGpu

	// TODO: add rbac
	//AllowedRoles []string `bson:"allowedRoles,omitempty"`
}

type GpuClaimUpdateParams struct {
	Name *string
	Zone *string

	Requested *map[string]RequestedGpu `bson:"requested"`

	// TODO: add rbac
	//AllowedRoles []string `bson:"allowedRoles,omitempty"`
}
