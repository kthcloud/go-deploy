package model

type UserCreateParams struct {
	Username      string
	FirstName     string
	LastName      string
	Email         string
	IsAdmin       bool
	EffectiveRole *EffectiveRole
}

type UserUpdateParams struct {
	PublicKeys *[]PublicKey
	UserData   *[]UserData
}
