package deployment

type Port struct {
	Name string `json:"name" bson:"name"`
	Port int    `json:"port" bson:"port"`
}

type DeploymentUpdate struct {
	Ports []Port `json:"ports" bson:"ports"`
}
