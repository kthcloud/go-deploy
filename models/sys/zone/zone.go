package zone

const (
	ZoneTypeDeployment = "deployment"
	ZoneTypeVM         = "vm"
)

type Zone struct {
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Type        string  `bson:"type"`
	Interface   *string `bson:"interface"`
}
