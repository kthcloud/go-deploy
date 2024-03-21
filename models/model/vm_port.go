package model

type VmPortLease struct {
	VmID        string `bson:"vmId"`
	PrivatePort int    `bson:"privatePort"`
}

type VmPort struct {
	PublicPort int          `bson:"publicPort"`
	Zone       string       `bson:"zone"`
	Lease      *VmPortLease `bson:"lease"`
}
