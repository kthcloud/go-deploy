package vm

const (
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
)

type CreateParams struct {
	Name    string
	Version string

	Zone           string
	DeploymentZone *string

	NetworkID *string

	SshPublicKey string
	PortMap      map[string]PortCreateParams

	CpuCores int
	RAM      int
	DiskSize int
}

type UpdateParams struct {
	Name       *string
	OwnerID    *string
	SnapshotID *string
	PortMap    *map[string]PortUpdateParams
	CpuCores   *int
	RAM        *int

	// update owner
	TransferCode   *string
	TransferUserID *string
}

type ActionParams struct {
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
