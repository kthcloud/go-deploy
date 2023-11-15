package role

type Quotas struct {
	Deployments      int     `yaml:"deployments" structs:"deployments"`
	CpuCores         int     `yaml:"cpuCores" structs:"cpuCores"`
	RAM              int     `yaml:"ram" structs:"ram"`
	DiskSize         int     `yaml:"diskSize" structs:"diskSize"`
	Snapshots        int     `yaml:"snapshots" structs:"snapshots"`
	GpuLeaseDuration float64 `yaml:"gpuLeaseDuration" structs:"gpuLeaseDuration"` // in hours
}

type Permissions struct {
	ChooseZone        bool `yaml:"chooseZone" structs:"chooseZone"`
	ChooseGPU         bool `yaml:"chooseGpu" structs:"chooseGpu"`
	UseCustomDomains  bool `yaml:"useCustomDomains" structs:"useCustomDomains"`
	UseGPUs           bool `yaml:"useGpus" structs:"useGpus"`
	UsePrivilegedGPUs bool `yaml:"usePrivilegedGpus" structs:"usePrivilegedGpus"`
}

type Role struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	IamGroup    string      `yaml:"iamGroup"`
	Permissions Permissions `yaml:"permissions"`
	Quotas      Quotas      `yaml:"quotas"`
}
