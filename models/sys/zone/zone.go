package zone

const (
	// TypeDeployment is a zone type for deployments.
	TypeDeployment = "deployment"
	// TypeVM is a zone type for VMs.
	TypeVM = "vm"
)

type Zone struct {
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Type        string  `bson:"type"`
	Interface   *string `bson:"interface"`
}
