package deployment

type DeploymentUpdate struct {
	Envs []map[string]string `json:"envs"`
}
