package gpu_claim_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage GPU claims in the database.
type Client struct {
	base_clients.ResourceClient[model.GpuClaim]
	base_clients.ActivityResourceClient[model.GpuClaim]
}

// New returns a new GPU claim client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.GpuClaim]{
			Collection:     db.DB.GetCollection("gpuClaims"),
			IncludeDeleted: false,
		},
		ActivityResourceClient: base_clients.ActivityResourceClient[model.GpuClaim]{
			Collection: db.DB.GetCollection("gpuClaims"),
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithActivities adds a filter to the client to only include storage managers that does
// at least one of the given activities.
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
		Key: "$or", Value: orFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithNoActivities adds a filter to the client to only include storage managers without any activities.
func (client *Client) WithNoActivities() *Client {
	filter := bson.D{{
		Key: "activities", Value: bson.M{
			"$gte": bson.M{},
		},
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithRoles adds a filter to only get GpuClaims where there are no role requirements
// or the allowedRoles contains any of the provided role values.
func (client *Client) WithRoles(roles []string) *Client {
	filter := bson.D{
		{
			Key: "$or",
			Value: bson.A{
				bson.M{"allowedRoles": bson.M{"$exists": false}},
				bson.M{"allowedRoles": bson.M{"$size": 0}},
				bson.M{"allowedRoles": bson.M{"$in": roles}},
			},
		},
	}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// WithNames adds a filter to match name to one of the provided names
func (client *Client) WithNames(names []string) *Client {
	filter := bson.D{
		{
			Key: "name",
			Value: bson.M{
				"$in": names,
			},
		},
	}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// IncludeDeletedResources makes the client include deleted gpu claims.
func (client *Client) IncludeDeletedResources() *Client {
	client.ResourceClient.IncludeDeleted = true

	return client
}

// WithZone adds a filter to the client to only include groups with the given zone.
func (client *Client) WithZone(zone string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "zone", Value: zone}})

	return client
}
