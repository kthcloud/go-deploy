package zone

const (
	ZoneTypeDeployment = "deployment"
	ZoneTypeVM         = "vm"
)

type Zone struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Type        string `json:"type" bson:"type"`
}
