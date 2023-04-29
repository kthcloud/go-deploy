package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"strconv"
)

type PortForwardingRulePublic struct {
	ID          string `bson:"id"`
	VmID        string `bson:"vmId"`
	ProjectID   string `bson:"projectID"`
	NetworkID   string `bson:"networkId"`
	IpAddressID string `bson:"ipAddressId"`
	PublicPort  int    `bson:"publicPort"`
	PrivatePort int    `bson:"privatePort"`
	Protocol    string `bson:"protocol"`
}

func CreatePortForwardingRulePublicFromGet(rule *cloudstack.PortForwardingRule, projectID string) *PortForwardingRulePublic {
	publicPort, _ := strconv.Atoi(rule.Publicport)
	privatePort, _ := strconv.Atoi(rule.Privateport)

	return &PortForwardingRulePublic{
		ID:          rule.Id,
		ProjectID:   projectID,
		NetworkID:   rule.Networkid,
		IpAddressID: rule.Ipaddressid,
		VmID:        rule.Virtualmachineid,
		PublicPort:  publicPort,
		PrivatePort: privatePort,
		Protocol:    rule.Protocol,
	}
}
