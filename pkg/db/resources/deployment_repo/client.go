package deployment_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// Client is used to manage deployments in the database.
// It uses both the ResourceClient and ActivityResourceClient to provide a full set of operations.
type Client struct {
	Collection     *mongo.Collection
	RestrictUserID *string

	base_clients.ActivityResourceClient[model.Deployment]
	base_clients.ResourceClient[model.Deployment]
}

// New returns a new deployment client.
func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("deployments"),

		ActivityResourceClient: base_clients.ActivityResourceClient[model.Deployment]{
			Collection: db.DB.GetCollection("deployments"),
		},
		ResourceClient: base_clients.ResourceClient[model.Deployment]{
			Collection:     db.DB.GetCollection("deployments"),
			IncludeDeleted: false,
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	client.ActivityResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// ExcludeIDs adds a filter to the client to exclude the given IDs.
func (client *Client) ExcludeIDs(ids ...string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"id", bson.D{{"$nin", ids}}}})
	client.ActivityResourceClient.AddExtraFilter(bson.D{{"id", bson.D{{"$nin", ids}}}})

	return client
}

// IncludeDeletedResources makes the client include deleted storage deployments.
func (client *Client) IncludeDeletedResources() *Client {
	client.IncludeDeleted = true

	return client
}

// WithOwner adds a filter to the client to only include deployments with the given owner ID.
func (client *Client) WithOwner(ownerID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"ownerId", ownerID}})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter
	client.RestrictUserID = &ownerID

	return client
}

// WithActivities adds a filter to the client to only include deployments with the given activities.
func (client *Client) WithActivities(activities ...string) *Client {
	andFilter := bson.A{}

	for _, activity := range activities {
		andFilter = append(andFilter, bson.M{
			"activities." + activity: bson.M{
				"$exists": true,
			},
		})
	}

	filter := bson.D{{
		"$or", andFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithoutActivities adds a filter to the client to only include deployments without the given activities.
func (client *Client) WithoutActivities(activities ...string) *Client {
	andFilter := bson.A{}

	for _, activity := range activities {
		andFilter = append(andFilter, bson.M{
			"activities." + activity: bson.M{
				"$exists": false,
			},
		})
	}

	filter := bson.D{{
		"$and", andFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithNoActivities adds a filter to the client to only include deployments without any activities.
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

// WithNameRegex adds a filter to the client to only include deployments with a name matching the given regex.
func (client *Client) WithNameRegex(name string) *Client {
	filter := bson.D{{"name", bson.D{{"$regex", name}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// OlderThan adds a filter to the client to only include deployments created before the given timestamp.
func (client *Client) OlderThan(timestamp time.Time) *Client {
	filter := bson.D{{"createdAt", bson.D{{"$lt", timestamp}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithPendingCustomDomain adds a filter to the client to only include deployments with a pending custom domain.
func (client *Client) WithPendingCustomDomain() *Client {
	filter := bson.D{{"apps.main.customDomain.status", bson.D{{"$ne", model.CustomDomainStatusActive}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithZone adds a filter to the client to only include deployments in the given zone.
func (client *Client) WithZone(zone ...string) *Client {
	filter := bson.D{{"zone", bson.D{{"$in", zone}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// LastAccessedBefore adds a filter to the client to only include deployments that were last accessed before the given timestamp.
func (client *Client) LastAccessedBefore(timestamp time.Time) *Client {
	filter := bson.D{{"accessedAt", bson.D{{"$lt", timestamp}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// Enabled adds a filter to the client to only include deployments that are enabled.
func (client *Client) Enabled() *Client {
	filter := bson.D{{"apps.main.replicas", bson.D{{"$gt", 0}}}}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// Disabled adds a filter to the client to only include deployments that are disabled.
func (client *Client) Disabled() *Client {
	filter := bson.D{{"apps.main.replicas", 0}}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}
