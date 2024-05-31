package host_api

import (
	"fmt"
	"go-deploy/utils/requestutils"
)

func (c *Client) GetCapacities() (*Capacities, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host capacities. details: %s", err)
	}

	response, err := requestutils.DoJsonGetRequest[Capacities](c.URL+"/capacities", nil)
	if err != nil {
		return nil, makeError(err)
	}

	return response, nil
}
