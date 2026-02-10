package model

type Permissions struct {
	ChooseZone        bool `yaml:"chooseZone" structs:"chooseZone"`
	ChooseGPU         bool `yaml:"chooseGpu" structs:"chooseGpu"`
	UseCustomDomains  bool `yaml:"useCustomDomains" structs:"useCustomDomains"`
	UseGPUs           bool `yaml:"useGpus" structs:"useGpus"`
	UsePrivilegedGPUs bool `yaml:"usePrivilegedGpus" structs:"usePrivilegedGpus"`
	// If non nil then use explicit value, if nil then we allow use of vms (for backward compatibilty).
	// Only checked when creating a new VM, if a user already has a VM then we allow them to update it.
	UseVms *bool `yaml:"useVms,omitempty" structs:"useVms,omitempty"`
}
