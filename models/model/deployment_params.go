package model

type DeploymentCreateParams struct {
	Name string
	Type string

	CpuCores float64
	RAM      float64
	Replicas int

	Image        string
	InternalPort int
	Envs         []DeploymentEnv
	Volumes      []DeploymentVolume
	InitCommands []string
	Args         []string
	PingPath     string
	CustomDomain *string
	Visibility   Visibility

	Zone string
}

type DeploymentUpdateParams struct {
	Name    *string
	OwnerID *string

	CpuCores *float64
	RAM      *float64

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
