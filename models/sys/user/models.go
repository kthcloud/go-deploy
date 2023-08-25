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

type EffectiveRole struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

type User struct {
	ID            string        `json:"id" bson:"id"`
	Username      string        `json:"username" bson:"username"`
	Email         string        `json:"email" bson:"email"`
	IsAdmin       bool          `json:"isAdmin" bson:"isAdmin"`
	EffectiveRole EffectiveRole `json:"effectiveRole" bson:"effectiveRole"`
	PublicKeys    []PublicKey   `json:"publicKeys" bson:"publicKeys"`
}

type UserUpdate struct {
	Username   *string      `json:"username" bson:"username"`
	PublicKeys *[]PublicKey `json:"publicKeys" bson:"publicKeys"`
}
