package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"strconv"
)

type PortForwardingRulePublic struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`

	VmID        string `bson:"vmId"`
	IpAddressID string `bson:"ipAddressId"`
	PublicPort  int    `bson:"publicPort"`
	PrivatePort int    `bson:"privatePort"`
	Protocol    string `bson:"protocol"`
	Tags        []Tag  `bson:"tags"`
}

func CreatePortForwardingRulePublicFromGet(rule *cloudstack.PortForwardingRule) *PortForwardingRulePublic {
	publicPort, _ := strconv.Atoi(rule.Publicport)
	privatePort, _ := strconv.Atoi(rule.Privateport)

	tags := FromCsTags(rule.Tags)

	var name string
	for _, tag := range tags {
		if tag.Key == "deployName" {
			name = tag.Value
		}
	}

	return &PortForwardingRulePublic{
		ID:          rule.Id,
		Name:        name,
		IpAddressID: rule.Ipaddressid,
		VmID:        rule.Virtualmachineid,
		PublicPort:  publicPort,
		PrivatePort: privatePort,
		Protocol:    rule.Protocol,
		Tags:        tags,
	}
}
