package vm_port_repo

import (
	"context"
	"errors"
	"go-deploy/models/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// PortNotFoundErr is returned when a port is not found.
	PortNotFoundErr = errors.New("port not found")
	// NoPortsAvailableErr is returned when no ports are available.
	// This is likely due to the port range being full.
	NoPortsAvailableErr = errors.New("port not found")
)

// CreateIfNotExists creates the given port range if it does not already exist.
func (client *Client) CreateIfNotExists(publicPortStart, publicPortEnd int, zone string) (int, error) {
	toInsert := make([]interface{}, publicPortEnd-publicPortStart)
	for i := range toInsert {
		toInsert[i] = model.VmPort{
			PublicPort: publicPortStart + i,
			Zone:       zone,
			Lease:      nil,
		}
	}

	// We have a unique index for public port + zone, so it is safe to try to insert many at once
	res, err := client.Collection.InsertMany(context.Background(), toInsert, options.InsertMany().SetOrdered(false))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if res != nil {
				return len(res.InsertedIDs), nil
			}

			return 0, nil
		}

		return 0, err
	}

	return len(res.InsertedIDs), nil
}

// Lease leases a port for the given VM.
func (client *Client) Lease(publicPort, privatePort int, vmID, zone string) (*model.VmPort, error) {
	filter := bson.D{
		{"publicPort", publicPort},
		{"zone", zone},
		{"lease", nil},
	}

	update := bson.D{{"$set", bson.D{{"lease", model.VmPortLease{
		VmID:        vmID,
		PrivatePort: privatePort,
	}}}}}

	res := client.Collection.FindOneAndUpdate(context.Background(), filter, update)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return nil, PortNotFoundErr
		}

		return nil, res.Err()
	}

	var port model.VmPort
	err := res.Decode(&port)
	if err != nil {
		return nil, err
	}

	return &port, nil
}

// GetOrLeaseAny gets a port that is not leased, or leases a port for the given VM.
func (client *Client) GetOrLeaseAny(privatePort int, vmID, zone string) (*model.VmPort, error) {
	// First check if the lease already exists
	filter := bson.D{
		{"lease.vmId", vmID},
		{"lease.privatePort", privatePort},
	}

	vmPort, err := client.GetWithFilterAndProjection(filter, nil)
	if err != nil {
		return nil, err
	}

	if vmPort != nil {
		return vmPort, nil
	}

	// Fetch a port that is not leased
	filter = bson.D{
		{"zone", zone},
		{"lease", nil},
	}

	update := bson.D{
		{
			"$set", bson.D{
				{"lease", model.VmPortLease{
					VmID:        vmID,
					PrivatePort: privatePort,
				}},
			},
		},
	}

	res := client.Collection.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return nil, NoPortsAvailableErr
		}

		return nil, res.Err()
	}

	var port model.VmPort
	err = res.Decode(&port)
	if err != nil {
		return nil, err
	}

	return &port, nil
}

// ReleaseAll releases all ports leased by the given VM.
func (client *Client) ReleaseAll(vmID string) error {
	filter := bson.D{
		{"lease.vmId", vmID},
	}

	update := bson.D{
		{
			"$set", bson.D{
				{"lease", nil},
			},
		},
	}

	_, err := client.Collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// Erase erases a port.
// This removes the port from the database entirely.
func (client *Client) Erase(publicPort int, zone string) error {
	filter := bson.D{
		{"publicPort", publicPort},
		{"zone", zone},
	}

	_, err := client.Collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}

		return err
	}

	return nil
}
