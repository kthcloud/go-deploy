package model

type Permissions struct {
	ChooseZone        bool `yaml:"chooseZone" structs:"chooseZone"`
	ChooseGPU         bool `yaml:"chooseGpu" structs:"chooseGpu"`
	UseCustomDomains  bool `yaml:"useCustomDomains" structs:"useCustomDomains"`
	UseGPUs           bool `yaml:"useGpus" structs:"useGpus"`
	UsePrivilegedGPUs bool `yaml:"usePrivilegedGpus" structs:"usePrivilegedGpus"`
}
