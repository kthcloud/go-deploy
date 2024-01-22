package models

import "github.com/rancher/go-rancher/client"

type ClusterPublic struct {
	ID   string `bson:"id,omitempty"`
	Name string `bson:"name,omitempty"`
}

func CreateClusterPublicFromRead(cluster *client.Cluster) *ClusterPublic {
	return &ClusterPublic{
		ID:   cluster.Id,
		Name: cluster.Name,
	}
}
