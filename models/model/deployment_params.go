package model

type DeploymentCreateParams struct {
	Name string
	Type string

	CpuCores float64
	RAM      float64
	Replicas int

	Image        string
	InternalPort int
	Private      bool
	Envs         []DeploymentEnv
	Volumes      []DeploymentVolume
	InitCommands []string
	Args         []string
	PingPath     string
	CustomDomain *string

	Zone string
}

type DeploymentUpdateParams struct {
	Name    *string
	OwnerID *string

	CpuCores *float64
	RAM      *float64

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
}

type DeploymentUpdateOwnerParams struct {
	NewOwnerID    string
	OldOwnerID    string
	MigrationCode *string
}
