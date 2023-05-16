package conf

type Quota struct {
	Deployment int `json:"deployment"`
	CPU        int `json:"cpu"`
	Memory     int `json:"memory"`
	Disk       int `json:"disk"`
}

func (e *Environment) GetQuota(roles []string) *Quota {
	// this function should have logic to return the highest quota given the roles
	// right now it only checks if you are a power user role or not, and tries to find the quota for the power user role

	if len(roles) == 0 {
		return nil
	}

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
				Deployment: quota.Deployment,
				CPU:        quota.CPU,
				Memory:     quota.Memory,
				Disk:       quota.Disk,
			}
		}
	}

	return nil
}
