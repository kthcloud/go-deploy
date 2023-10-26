package user

type UpdateParams struct {
	Username   *string      `json:"username" bson:"username"`
	PublicKeys *[]PublicKey `json:"publicKeys" bson:"publicKeys"`
	Onboarded  *bool        `json:"onboarded" bson:"onboarded"`
}
