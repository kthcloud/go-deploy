package tmp_zone_repo

import (
	"context"
	"go-deploy/models/model"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Register(zone *model.Zone) error {
	// Register zone
	// Check if it already exists, if so, update it
	// Otherwise, insert it
	zone, err := client.GetByName(zone.Name)
	if err != nil {
		return err
	}

	if zone == nil {
		// Zone does not exist, insert it
		_, err = client.Collection.InsertOne(context.TODO(), zone)
		return err
	}

	// Zone exists, update it
	// Right now zone does not have any fields to update, but this is a placeholder for future updates
	update := bson.D{}
	return client.SetWithBsonByName(zone.Name, update)
}
