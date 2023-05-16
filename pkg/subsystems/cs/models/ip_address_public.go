package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"net"
)

type PublicIpAddressPublic struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`

	IpAddress net.IP `bson:"ipAddress"`
	Tags      []Tag  `bson:"tags"`
}

func CreatePublicIpAddressPublicFromGet(ipAddress *cloudstack.PublicIpAddress) *PublicIpAddressPublic {
	tags := FromCsTags(ipAddress.Tags)

	var name string
	for _, tag := range tags {
		if tag.Key == "deployName" {
			name = tag.Value
		}
	}

	return &PublicIpAddressPublic{
		ID:        ipAddress.Id,
		Name:      name,
		IpAddress: net.ParseIP(ipAddress.Ipaddress),
		Tags:      tags,
	}
}
