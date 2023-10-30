package user

type UpdateParams struct {
	PublicKeys *[]PublicKey `json:"publicKeys" bson:"publicKeys"`
	Onboarded  *bool        `json:"onboarded" bson:"onboarded"`
}
