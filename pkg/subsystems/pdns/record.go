package pdns

import (
	"go-deploy/pkg/subsystems/pdns/models"
)

func (client *Client) ReadRecord(id string) (*models.RecordPublic, error) {
	return nil, nil
}

func (client *Client) CreateRecord(public *models.RecordPublic) (string, error) {

	return "", nil
}

func (client *Client) DeleteRecord(id string) error {
	return nil
}
