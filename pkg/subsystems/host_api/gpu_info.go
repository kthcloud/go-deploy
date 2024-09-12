package host_api

import (
	"fmt"
	"github.com/kthcloud/go-deploy/utils/requestutils"
)

func (c *Client) GetGpuInfo() ([]GpuInfo, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host gpu info. details: %s", err)
	}

	response, err := requestutils.DoJsonGetRequest[[]GpuInfo](c.URL+"/gpuInfo", nil)
	if err != nil {
		return nil, makeError(err)
	}

	return *response, nil
}
