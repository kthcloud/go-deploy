package user

type PublicKey struct {
	Name string `json:"name" bson:"name"`
	Key  string `json:"key" bson:"key"`
}

type Quota struct {
	Deployment int `json:"deployment" bson:"deployment"`
	CPU        int `json:"cpu" bson:"cpu"`
	Memory     int `json:"memory" bson:"memory"`
	Disk       int `json:"disk" bson:"disk"`
}

type User struct {
	ID         string      `json:"id" bson:"id"`
	Username   string      `json:"username" bson:"username"`
	Email      string      `json:"email" bson:"email"`
	Roles      []string    `json:"roles" bson:"roles"`
	PublicKeys []PublicKey `json:"publicKeys" bson:"publicKeys"`
}

type UserUpdate struct {
	Username        *string      `json:"username" bson:"username"`
	Email           *string      `json:"email" bson:"email"`
	VmQuota         *int         `json:"vmQuota" bson:"vmQuota"`
	DeploymentQuota *int         `json:"deploymentQuota" bson:"deploymentQuota"`
	PublicKeys      *[]PublicKey `json:"publicKeys" bson:"publicKeys"`
}
