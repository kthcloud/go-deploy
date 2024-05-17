package model

const (
	ActionStart            = "start"
	ActionStop             = "stop"
	ActionRestart          = "restart"
	ActionRestartIfRunning = "restartIfRunning"
)

type VmCreateParams struct {
	Name string
	Zone string

	SshPublicKey string
	PortMap      map[string]PortCreateParams

	CpuCores int
	RAM      int
	DiskSize int
}

type VmUpdateParams struct {
	Name       *string
	OwnerID    *string
	SnapshotID *string
	PortMap    *map[string]PortUpdateParams
	CpuCores   *int
	RAM        *int
}

type VmUpdateOwnerParams struct {
	NewOwnerID string
	OldOwnerID string
}

type VmActionParams struct {
	Action string
}

type PortCreateParams struct {
	Name      string
	Port      int
	Protocol  string
	HttpProxy *HttpProxyCreateParams
}

type PortUpdateParams struct {
	Name      string
	Port      int
	Protocol  string
	HttpProxy *HttpProxyUpdateParams
}

type HttpProxyCreateParams struct {
	Name         string
	CustomDomain *string
}

type HttpProxyUpdateParams struct {
	Name         string
	CustomDomain *string
}

type CreateSnapshotParams struct {
	Name        string `bson:"name"`
	UserCreated bool   `bson:"userCreated"`
	Overwrite   bool   `bson:"overwrite"`
}
