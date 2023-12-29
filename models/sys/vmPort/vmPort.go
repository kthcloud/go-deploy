package vmPort

type Lease struct {
	VmID        string `bson:"vmId"`
	UserID      string `bson:"userId"`
	PrivatePort int    `bson:"privatePort"`
}

type VmPort struct {
	PublicPort int    `bson:"publicPort"`
	Zone       string `bson:"zone"`
	Lease      *Lease `bson:"lease"`
}
