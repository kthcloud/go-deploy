package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"strconv"
	"time"
)

type PortForwardingRulePublic struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdAt"`

	VmID        string `bson:"vmId"`
	NetworkID   string `bson:"networkId"`
	IpAddressID string `bson:"ipAddressId"`

	PublicPort  int    `bson:"publicPort"`
	PrivatePort int    `bson:"privatePort"`
	Protocol    string `bson:"protocol"`

	Tags []Tag `bson:"tags"`
}

func (rule *PortForwardingRulePublic) Created() bool {
	return rule.ID != ""
}

func CreatePortForwardingRulePublicFromGet(rule *cloudstack.PortForwardingRule) *PortForwardingRulePublic {
	publicPort, _ := strconv.Atoi(rule.Publicport)
	privatePort, _ := strconv.Atoi(rule.Privateport)

	tags := FromCsTags(rule.Tags)

	var name string
	var createdAt time.Time

	for _, tag := range tags {
		if tag.Key == "name" {
			name = tag.Value
		}

		if tag.Key == "createdAt" {
			createdAt, _ = time.Parse(time.RFC3339, tag.Value)
		}
	}

	return &PortForwardingRulePublic{
		ID:        rule.Id,
		Name:      name,
		CreatedAt: createdAt,

		VmID:        rule.Virtualmachineid,
		NetworkID:   rule.Networkid,
		IpAddressID: rule.Ipaddressid,

		PublicPort:  publicPort,
		PrivatePort: privatePort,
		Protocol:    rule.Protocol,

		Tags: tags,
	}
}
