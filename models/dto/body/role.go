package body

type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Quota       *Quota   `json:"quota,omitempty"`
}
