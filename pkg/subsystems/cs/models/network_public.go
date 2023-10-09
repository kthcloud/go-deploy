package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type NetworkPublic struct {
	ID          string    `bson:"id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"createdAt"`
	Tags        []Tag     `bson:"tags"`
}

func (n *NetworkPublic) Created() bool {
	return n.ID != ""
}

func (n *NetworkPublic) IsPlaceholder() bool {
	return false
}

func CreateNetworkPublicFromGet(network *cloudstack.Network) *NetworkPublic {
	tags := FromCsTags(network.Tags)

	var name string
	for _, tag := range tags {
		if tag.Key == "deployName" {
			name = tag.Value
		}
	}

	return &NetworkPublic{
		ID:          network.Id,
		Name:        name,
		Description: network.Displaytext,
		CreatedAt:   formatCreatedAt(network.Created),
	}
}

func CreateNetworkPublicFromCreate(network *cloudstack.CreateNetworkResponse) *NetworkPublic {
	return CreateNetworkPublicFromGet(
		&cloudstack.Network{
			Id:          network.Id,
			Name:        network.Name,
			Displaytext: network.Displaytext,
			Created:     network.Created,
		},
	)
}

func CreateNetworkPublicFromUpdate(network *cloudstack.UpdateNetworkResponse) *NetworkPublic {
	return CreateNetworkPublicFromGet(
		&cloudstack.Network{
			Id:          network.Id,
			Name:        network.Name,
			Displaytext: network.Displaytext,
			Created:     network.Created,
		},
	)
}
