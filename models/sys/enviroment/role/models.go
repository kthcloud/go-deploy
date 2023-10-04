package role

type Quotas struct {
	Deployments int `yaml:"deployments"`
	CpuCores    int `yaml:"cpuCores"`
	RAM         int `yaml:"ram"`
	DiskSize    int `yaml:"diskSize"`
	Snapshots   int `yaml:"snapshots"`
}

type Permissions struct {
	ChooseZone        bool    `yaml:"chooseZone" structs:"chooseZone"`
	ChooseGPU         bool    `yaml:"chooseGpu" structs:"chooseGpu"`
	UseCustomDomain   bool    `yaml:"useCustomDomain" structs:"useCustomDomain"`
	UseGPUs           bool    `yaml:"useGpus" structs:"useGpus"`
	UsePrivilegedGPUs bool    `yaml:"usePrivilegedGpus" structs:"usePrivilegedGpus"`
	GpuLeaseDuration  float64 `yaml:"gpuLeaseDuration" structs:"gpuLeaseDuration"` // in hours
}

type Role struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	IamGroup    string      `yaml:"iamGroup"`
	Permissions Permissions `yaml:"permissions"`
	Quotas      Quotas      `yaml:"quotas"`
}
