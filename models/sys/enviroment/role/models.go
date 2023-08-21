package role

type Quotas struct {
	Deployments int `yaml:"deployments"`
	CpuCores    int `yaml:"cpuCores"`
	RAM         int `yaml:"ram"`
	DiskSize    int `yaml:"diskSize"`
	Snapshots   int `yaml:"snapshots"`
}

type Role struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	IamGroup    string `yaml:"iamGroup"`
	Permissions struct {
		ChooseZone        bool `yaml:"chooseZone"`
		ChooseGPU         bool `yaml:"chooseGpu"`
		UseGPUs           bool `yaml:"useGpus"`
		UsePrivilegedGPUs bool `yaml:"seePrivilegedGpus"`
		// in hours
		GpuLeaseDuration float64 `yaml:"gpuLeaseDuration"`
	}
	Quotas Quotas `yaml:"quotas"`
}
