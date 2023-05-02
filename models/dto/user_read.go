package dto

type UserRead struct {
	ID         string   `json:"id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Admin      bool     `json:"admin"`
	PowerUser  bool     `json:"powerUser"`
	PublicKeys []string `json:"publicKeys"`
}
