package storage_manager

type CreateParams struct {
	UserID string `json:"userId"`
	Zone   string `json:"zone,omitempty"`
}
