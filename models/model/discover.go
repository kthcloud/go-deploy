package model

import (
	body2 "go-deploy/dto/v2/body"
)

type Discover struct {
	Version string
	Roles   []Role
}

func (d *Discover) ToDTO() body2.DiscoverRead {
	roles := make([]body2.Role, len(d.Roles))
	for i, r := range d.Roles {
		roles[i] = r.ToDTO(true)
	}

	return body2.DiscoverRead{
		Version: d.Version,
		Roles:   roles,
	}
}
