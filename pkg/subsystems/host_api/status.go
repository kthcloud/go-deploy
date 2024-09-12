package host_api

import (
	"fmt"
	"github.com/kthcloud/go-deploy/utils/requestutils"
)

func (c *Client) GetStatus() (*Status, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host status. details: %s", err)
	}

	response, err := requestutils.DoJsonGetRequest[Status](c.URL+"/status", nil)
	if err != nil {
		return nil, makeError(err)
	}

	return response, nil
}
