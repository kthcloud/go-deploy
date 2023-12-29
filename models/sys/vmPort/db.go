package vmPort

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
