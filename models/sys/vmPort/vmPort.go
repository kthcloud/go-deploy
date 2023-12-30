package vmPort

type Lease struct {
	VmID        string `bson:"vmId"`
	PrivatePort int    `bson:"privatePort"`
}

type VmPort struct {
	PublicPort int    `bson:"publicPort"`
	Zone       string `bson:"zone"`
	Lease      *Lease `bson:"lease"`
}
