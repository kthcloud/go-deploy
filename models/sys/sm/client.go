package sm

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/activityResource"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection      *mongo.Collection
	RestrictOwnerID *string

	activityResource.ActivityResourceClient[SM]
	resource.ResourceClient[SM]
}

func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("storageManagers"),

		ActivityResourceClient: activityResource.ActivityResourceClient[SM]{
			Collection: db.DB.GetCollection("storageManagers"),
		},
		ResourceClient: resource.ResourceClient[SM]{
			Collection:     db.DB.GetCollection("storageManagers"),
			IncludeDeleted: false,
		},
	}
}

func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	client.ActivityResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) IncludeDeletedResources() *Client {
	client.IncludeDeleted = true

	return client
}

func (client *Client) RestrictToOwner(ownerID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"ownerId", ownerID}})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter
	client.RestrictOwnerID = &ownerID

	return client
}

func (client *Client) WithActivities(activities ...string) *Client {
	orFilter := bson.A{}

	for _, activity := range activities {
		orFilter = append(orFilter, bson.M{
			"activities." + activity: bson.M{
				"$exists": true,
			},
		})
	}

	filter := bson.D{{
		"$or", orFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) WithNoActivities() *Client {
	filter := bson.D{{
		"activities", bson.M{
			"$gte": bson.M{},
		},
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}
