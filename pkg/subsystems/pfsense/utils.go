package pfsense

import (
	"fmt"
	"go-deploy/pkg/subsystems/pfsense/models"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func makeApiError(pfsenseResponse *models.Response, makeError func(error) error) error {
	resCode := pfsenseResponse.Code
	resMsg := pfsenseResponse.Message
	errorMessage := fmt.Sprintf("erroneous request (%d). details: %s", resCode, resMsg)
	return makeError(fmt.Errorf(errorMessage))
}

func (client *Client) getFreePublicPort() int {
	start := client.portRangeStart
	end := client.portRangeEnd

	port := start + rand.Int()%(end-start)
	return port
}
