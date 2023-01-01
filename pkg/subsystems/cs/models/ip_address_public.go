package models

import (
	"github.com/apache/cloudstack-go/v2/cloudstack"
	"net"
)

type PublicIpAddressPublic struct {
	ID        string `bson:"id"`
	NetworkID string `bson:"networkId"`
	ZoneID    string `bson:"zoneId"`
	ProjectID string `bson:"projectId"`
	IpAddress net.IP `bson:"ipAddress"`
}

func CreatePublicIpAddressPublicFromGet(ipAddress *cloudstack.PublicIpAddress) *PublicIpAddressPublic {
	return &PublicIpAddressPublic{
		ID:        ipAddress.Id,
		NetworkID: ipAddress.Networkid,
		ZoneID:    ipAddress.Zoneid,
		ProjectID: ipAddress.Projectid,
		IpAddress: net.ParseIP(ipAddress.Ipaddress),
	}
}
