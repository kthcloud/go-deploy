package models

import "go-deploy/pkg/imp/cloudstack"

type HostPublic struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	State         string `json:"state"`
	ResourceState string `json:"resourceState"`
}

func CreatePublicFromGet(host *cloudstack.Host) *HostPublic {
	return &HostPublic{
		ID:            host.Id,
		Name:          host.Name,
		State:         host.State,
		ResourceState: host.Resourcestate,
	}
}
