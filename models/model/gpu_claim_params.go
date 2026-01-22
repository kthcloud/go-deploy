package model

type GpuClaimCreateParams struct {
	Name string `json:"name" bson:"name"`
	Zone string `json:"zone" bson:"zone"`

	Requested []RequestedGpuCreate `json:"requested" bson:"requested"`

	AllowedRoles []string `bson:"allowedRoles,omitempty"`
}

type GpuClaimUpdateParams struct {
	Name *string `json:"name" bson:"name"`
	Zone *string `json:"zone" bson:"zone"`

	Requested *[]RequestedGpuCreate `json:"requested" bson:"requested"`

	AllowedRoles *[]string `bson:"allowedRoles,omitempty"`
}
