package model

type DeploymentCreateParams struct {
	Name string
	Type string

	CpuCores float64
	RAM      float64
	Replicas int

	Image         string
	InternalPort  int
	InternalPorts []int
	Envs          []DeploymentEnv
	Volumes       []DeploymentVolume
	InitCommands  []string
	Args          []string
	PingPath      string
	CustomDomain  *string
	Visibility    string

	NeverStale bool

	Zone string
}

type DeploymentUpdateParams struct {
	Name    *string
	OwnerID *string

	CpuCores *float64
	RAM      *float64

	Envs          *[]DeploymentEnv
	InternalPort  *int
	InternalPorts *[]int
	Volumes       *[]DeploymentVolume
	InitCommands  *[]string
	Args          *[]string
	CustomDomain  *string
	Image         *string
	PingPath      *string
	Replicas      *int
	Visibility    *string

	NeverStale *bool
}

type DeploymentUpdateOwnerParams struct {
	NewOwnerID    string
	OldOwnerID    string
	MigrationCode *string
}
