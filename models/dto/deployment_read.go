package dto

type DeploymentRead struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	Status string `json:"status"`
}
