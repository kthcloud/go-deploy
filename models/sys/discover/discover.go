package discover

import (
	"go-deploy/models/sys/role"
)

type Discover struct {
	Version string
	Roles   []role.Role
}
