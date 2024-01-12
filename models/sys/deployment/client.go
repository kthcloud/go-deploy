package deployment

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/activityResource"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Client struct {
	Collection     *mongo.Collection
	RestrictUserID *string

	activityResource.ActivityResourceClient[Deployment]
	resource.ResourceClient[Deployment]
}

func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("deployments"),

		ActivityResourceClient: activityResource.ActivityResourceClient[Deployment]{
			Collection: db.DB.GetCollection("deployments"),
		},
		ResourceClient: resource.ResourceClient[Deployment]{
			Collection:     db.DB.GetCollection("deployments"),
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

func (client *Client) ExcludeIDs(ids ...string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"id", bson.D{{"$nin", ids}}}})
	client.ActivityResourceClient.AddExtraFilter(bson.D{{"id", bson.D{{"$nin", ids}}}})

	return client
}

func (client *Client) IncludeDeletedResources() *Client {
	client.IncludeDeleted = true

	return client
}

func (client *Client) RestrictToOwner(ownerID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"ownerId", ownerID}})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter
	client.RestrictUserID = &ownerID

	return client
}

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
		"$and", andFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

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

func (client *Client) WithGitHubWebhookID(id int64) *Client {
	filter := bson.D{{"subsystems.github.webhook.id", id}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) WithNameRegex(name string) *Client {
	filter := bson.D{{"name", bson.D{{"$regex", name}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) OlderThan(timestamp time.Time) *Client {
	filter := bson.D{{"createdAt", bson.D{{"$lt", timestamp}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) WithPendingCustomDomain() *Client {
	filter := bson.D{{"apps.main.customDomain.status", bson.D{{"$ne", CustomDomainStatusActive}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}
