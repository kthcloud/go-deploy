package dto

type DeploymentRead struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerID"`
	Status  string `json:"status"`
}
