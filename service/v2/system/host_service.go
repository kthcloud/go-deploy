package system

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/host_repo"
)

// ListHosts gets a list of hosts
func (c *Client) ListHosts() ([]model.Host, error) {
	hosts, err := host_repo.New().Activated().List()
	if err != nil {
		return nil, err
	}

	return hosts, nil
}
