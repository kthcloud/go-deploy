package dto

type UserUpdate struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	PublicKeys []string `json:"publicKeys"`
}
