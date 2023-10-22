package landing

import (
	"fmt"
	"go-deploy/pkg/subsystems/landing/models"
	"go-deploy/utils/requestutils"
)

func (client *Client) ReadGpuInfo() (*models.GpuInfoRead, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gpu info. details: %w", err)
	}

	res, err := client.doRequest("GET", "/internal/gpuInfo")
	if err != nil {
		return nil, makeError(err)
	}

	// check if good request
	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return nil, makeError(fmt.Errorf("bad status code: %d", res.StatusCode))
	}

	var gpus []models.GpuInfoRead
	err = requestutils.ParseBody(res.Body, &gpus)
	if err != nil {
		return nil, makeError(err)
	}

	return &gpus[0], nil
}
