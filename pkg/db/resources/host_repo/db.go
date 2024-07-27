package host_repo

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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
		{"displayName", host.DisplayName},
		{"zone", host.Zone},

		{"ip", host.IP},
		{"port", host.Port},

		{"lastSeenAt", time.Now()},

		{"enabled", host.Enabled},
		{"schedulable", host.Schedulable},
	}

	err = client.SetWithBsonByName(host.Name, update)
	if err != nil {
		return fmt.Errorf("failed to update host %s. details: %w", host.Name, err)
	}

	return nil
}

func (client *Client) MarkSchedulable(name string, schedulable bool) error {
	return client.SetWithBsonByName(name, bson.D{{"schedulable", schedulable}})
}

func (client *Client) DeactivateHost(name string, until ...time.Time) error {
	var deactivatedUntil time.Time
	if len(until) > 0 {
		deactivatedUntil = until[0]
	} else {
		deactivatedUntil = time.Now().AddDate(1000, 0, 0) // 1000 years is sort of forever ;)
	}

	return client.SetWithBsonByName(name, bson.D{{"deactivatedUntil", deactivatedUntil}})
}
