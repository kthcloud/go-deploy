package dto

type DeploymentRead struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerId"`
	Status  string `json:"status"`
	URL     string `json:"url"`
}
