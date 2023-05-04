package user

type PublicKey struct {
	Name string `json:"name" bson:"name"`
	Key  string `json:"key" bson:"key"`
}

type User struct {
	ID              string      `json:"id" bson:"id"`
	Username        string      `json:"username" bson:"username"`
	Email           string      `json:"email" bson:"email"`
	VmQuota         int         `json:"vmQuota" bson:"vmQuota"`
	DeploymentQuota int         `json:"deploymentQuota" bson:"deploymentQuota"`
	IsAdmin         bool        `json:"isAdmin" bson:"isAdmin"`
	IsPowerUser     bool        `json:"isPowerUser" bson:"isPowerUser"`
	PublicKeys      []PublicKey `json:"publicKeys" bson:"publicKeys"`
}

type UserUpdate struct {
	Username        *string      `json:"username" bson:"username"`
	Email           *string      `json:"email" bson:"email"`
	VmQuota         *int         `json:"vmQuota" bson:"vmQuota"`
	DeploymentQuota *int         `json:"deploymentQuota" bson:"deploymentQuota"`
	PublicKeys      *[]PublicKey `json:"publicKeys" bson:"publicKeys"`
}
