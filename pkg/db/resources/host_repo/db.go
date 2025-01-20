package host_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"go.mongodb.org/mongo-driver/bson"
)

func (client *Client) Register(host *model.Host) error {
	// Register host
	// Check if it already exists, if so, update it
	// Otherwise, insert it
	current, err := client.GetByName(host.Name)
	if err != nil {

		return err
	}

	if current == nil {
		// Host does not exist, insert it
		_, err = client.Collection.InsertOne(context.TODO(), host)
		if err != nil {
			return fmt.Errorf("failed to insert host %s. details: %w", host.Name, err)
		}

		return nil
	}

	// Host exists, update it
	update := bson.D{
		{Key: "displayName", Value: host.DisplayName},
		{Key: "zone", Value: host.Zone},

		{Key: "ip", Value: host.IP},
		{Key: "port", Value: host.Port},

		{Key: "lastSeenAt", Value: time.Now()},

		{Key: "enabled", Value: host.Enabled},
		{Key: "schedulable", Value: host.Schedulable},
	}

	err = client.SetWithBsonByName(host.Name, update)
	if err != nil {
		return fmt.Errorf("failed to update host %s. details: %w", host.Name, err)
	}

	return nil
}

func (client *Client) MarkSchedulable(name string, schedulable bool) error {
	return client.SetWithBsonByName(name, bson.D{{Key: "schedulable", Value: schedulable}})
}

func (client *Client) DeactivateHost(name string, until ...time.Time) error {
	var deactivatedUntil time.Time
	if len(until) > 0 {
		deactivatedUntil = until[0]
	} else {
		deactivatedUntil = time.Now().AddDate(1000, 0, 0) // 1000 years is sort of forever ;)
	}

	return client.SetWithBsonByName(name, bson.D{{Key: "deactivatedUntil", Value: deactivatedUntil}})
}
