package vmPort

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	PortNotFoundErr     = errors.New("port not found")
	NoPortsAvailableErr = errors.New("port not found")
)

func (client *Client) GetByPublicPort(publicPort int, zone string) (*VmPort, error) {
	filter := bson.D{
		{"publicPort", publicPort},
		{"zone", zone},
	}

	return client.GetWithFilterAndProjection(filter, nil)
}

func (client *Client) GetByLease(vmID string, privatePort int) (*VmPort, error) {
	filter := bson.D{
		{"lease.vmId", vmID},
		{"lease.privatePort", privatePort},
	}

	return client.GetWithFilterAndProjection(filter, nil)
}

func (client *Client) CreateIfNotExists(publicPortStart, publicPortEnd int, zone string) (int, error) {
	toInsert := make([]interface{}, publicPortEnd-publicPortStart)
	for i := range toInsert {
		toInsert[i] = VmPort{
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

func (client *Client) Lease(publicPort, privatePort int, vmID, zone string) (*VmPort, error) {
	filter := bson.D{
		{"publicPort", publicPort},
		{"zone", zone},
		{"lease", nil},
	}

	update := bson.D{{"$set", bson.D{{"lease", Lease{
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

	var port VmPort
	err := res.Decode(&port)
	if err != nil {
		return nil, err
	}

	return &port, nil
}

func (client *Client) GetOrLeaseAny(privatePort int, vmID, zone string) (*VmPort, error) {
	// First check if the lease already exists
	filter := bson.D{
		{"lease.vmId", vmID},
		{"lease.privatePort", privatePort},
	}

	vmPort, err := client.GetByLease(vmID, privatePort)
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
				{"lease", Lease{
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

	var port VmPort
	err = res.Decode(&port)
	if err != nil {
		return nil, err
	}

	return &port, nil
}

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

func (client *Client) Delete(publicPort int, zone string) error {
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
