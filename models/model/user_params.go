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

	// Onboarded
	// Deprecated: This field is deprecated and will be removed in the future.
	Onboarded *bool
}
