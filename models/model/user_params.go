package model

type UserSynchronizeParams struct {
	Username      string
	FirstName     string
	LastName      string
	Email         string
	IsAdmin       bool
	EffectiveRole *EffectiveRole
}

type UserUpdateParams struct {
	ApiKeys    *[]ApiKey
	PublicKeys *[]PublicKey
	UserData   *[]UserData
}
