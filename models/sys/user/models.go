package user

type PublicKey struct {
	Name string `json:"name" bson:"name"`
	Key  string `json:"key" bson:"key"`
}

type Usage struct {
	Deployments int `json:"deployments" bson:"deployments"`
	CpuCores    int `json:"cpuCores" bson:"cpuCores"`
	RAM         int `json:"ram" bson:"ram"`
	DiskSize    int `json:"diskSize" bson:"diskSize"`
	Snapshots   int `json:"snapshots" bson:"snapshots"`
}

type User struct {
	ID         string      `json:"id" bson:"id"`
	Username   string      `json:"username" bson:"username"`
	Email      string      `json:"email" bson:"email"`
	Roles      []string    `json:"roles" bson:"roles"`
	PublicKeys []PublicKey `json:"publicKeys" bson:"publicKeys"`
}

type UserUpdate struct {
	Username   *string      `json:"username" bson:"username"`
	Email      *string      `json:"email" bson:"email"`
	PublicKeys *[]PublicKey `json:"publicKeys" bson:"publicKeys"`
}
