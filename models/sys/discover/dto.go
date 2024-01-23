package discover

import "go-deploy/models/dto/v1/body"

// ToDTO converts a Discover to a body.DiscoverRead DTO.
func (d *Discover) ToDTO() *body.DiscoverRead {
	roles := make([]body.Role, len(d.Roles))
	for i, r := range d.Roles {
		roles[i] = r.ToDTO(true)
	}

	return &body.DiscoverRead{
		Version: d.Version,
		Roles:   roles,
	}
}
