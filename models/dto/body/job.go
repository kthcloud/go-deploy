package body

type JobRead struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userId"`
	Type      string  `json:"type"`
	Status    string  `json:"status"`
	LastError *string `json:"lastError,omitempty"`
}
