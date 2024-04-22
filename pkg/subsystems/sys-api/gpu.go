package sys_api

import (
	"fmt"
	"go-deploy/pkg/subsystems/sys-api/models"
	"go-deploy/utils/requestutils"
)

// ReadGpuInfo reads GPU info from the sys-api service.
func (client *Client) ReadGpuInfo() (*models.GpuInfoRead, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gpu info. details: %w", err)
	}

	if client.useMock {
		return getGpuInfoMock(), nil
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

func getGpuInfoMock() *models.GpuInfoRead {
	return &models.GpuInfoRead{
		GpuInfo: struct {
			Hosts []struct {
				Name string       `json:"name"`
				Zone string       `json:"zone"`
				GPUs []models.GPU `json:"gpus"`
			} `json:"hosts"`
		}{
			Hosts: []struct {
				Name string       `json:"name"`
				Zone string       `json:"zone"`
				GPUs []models.GPU `json:"gpus"`
			}{
				{
					Name: "go-deploy-fake-host",
					Zone: "local",
					GPUs: []models.GPU{
						{
							Name:     "Quadro RTX 5000",
							Slot:     "02:00.0",
							Vendor:   "NVIDIA Corporation",
							VendorID: "10de",
							Bus:      "02",
							DeviceID: "1eb0",
						},
					},
				},
			},
		},
	}
}
