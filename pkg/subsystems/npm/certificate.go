package npm

import (
	"fmt"
	"go-deploy/pkg/subsystems/npm/models"
	"go-deploy/utils/requestutils"
)

func (client *Client) GetWildcardCertificateID(parentDomain string) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get certificates. details: %s", err)
	}

	res, err := client.doRequest("GET", "/nginx/certificates")
	if err != nil {
		return -1, makeError(err)
	}

	// check if good request
	if !requestutils.IsGoodStatusCode(res.StatusCode) {
		return -1, makeApiError(res.Body, makeError)
	}

	var certificates []models.Certificate
	err = requestutils.ParseBody(res.Body, &certificates)
	if err != nil {
		return -1, makeError(err)
	}

	searchFor := fmt.Sprintf("*.%s", parentDomain)
	for _, certificate := range certificates {
		for _, domainName := range certificate.DomainNames {
			if domainName == searchFor {
				return certificate.ID, nil
			}
		}
	}

	err = makeError(fmt.Errorf("certificate list did not contain a certificate with domain name %s", searchFor))
	return -1, makeError(err)
}
