package model

type DeploymentCreateParams struct {
	Name string
	Type string

	Image        string
	InternalPort int
	Private      bool
	Envs         []DeploymentEnv
	Volumes      []DeploymentVolume
	InitCommands []string
	Args         []string
	PingPath     string
	CustomDomain *string
	Replicas     *int

	Zone string
}

type DeploymentUpdateParams struct {
	// Normal Update
	Name         *string
	OwnerID      *string
	Private      *bool
	Envs         *[]DeploymentEnv
	InternalPort *int
	Volumes      *[]DeploymentVolume
	InitCommands *[]string
	Args         *[]string
	CustomDomain *string
	Image        *string
	PingPath     *string
	Replicas     *int

	// Ownership update
	TransferUserID *string
	TransferCode   *string
}
