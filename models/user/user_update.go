package user

type UserUpdate struct {
	Username        *string   `json:"username" bson:"username"`
	Email           *string   `json:"email" bson:"email"`
	VmQuota         *int      `json:"vmQuota" bson:"vmQuota"`
	DeploymentQuota *int      `json:"deploymentQuota" bson:"deploymentQuota"`
	PublicKeys      *[]string `json:"publicKeys" bson:"publicKeys"`
}
