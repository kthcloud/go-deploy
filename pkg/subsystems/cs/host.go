package cs

import "go-deploy/pkg/subsystems/cs/models"

func (client *Client) ReadHostByName(name string) (*models.HostPublic, error) {
	host, _, err := client.CsClient.Host.GetHostByName(name)
	if err != nil {
		return nil, err
	}

	if host == nil {
		return nil, nil
	}

	return models.CreatePublicFromGet(host), nil
}
