package conf

type Quota struct {
	Deployments int `json:"deployments"`
	CpuCores    int `json:"cpuCores"`
	RAM         int `json:"ram"`
	DiskSize    int `json:"diskSize"`
}

func (e *Environment) GetQuota(roles []string) *Quota {
	// this function should have logic to return the highest quota given the roles
	// right now it only checks if you are a power user role or not, and tries to find the quota for the power user role

	for _, role := range roles {
		if role == Env.Keycloak.PowerUserGroup {
			quota := e.FindQuota(role)
			if quota != nil {
				return quota
			}
		}
	}

	defaultQuota := e.FindQuota("default")
	if defaultQuota != nil {
		return defaultQuota
	}

	return nil
}

func (e *Environment) FindQuota(role string) *Quota {
	for _, quota := range Env.Quotas {
		if quota.Role == role {
			return &Quota{
				Deployments: quota.Deployments,
				CpuCores:    quota.CpuCores,
				RAM:         quota.RAM,
				DiskSize:    quota.DiskSize,
			}
		}
	}

	return nil
}
