package models

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/log"
	"go-deploy/utils/subsystemutils"
	"net"
	"strconv"
)

type PortForwardingRulePublic struct {
	ID              string `bson:"id"`
	Name            string `bson:"name"`
	ExternalAddress net.IP `bson:"destinationAddress"`
	ExternalPort    int    `bson:"destinationPort"`
	LocalAddress    net.IP `bson:"localAddress"`
	LocalPort       int    `bson:"localPort"`
}

func (public *PortForwardingRulePublic) GetDescription() string {
	prefixedName := subsystemutils.GetPrefixedName(public.Name)
	return fmt.Sprintf("%s-%s", prefixedName, public.ID)
}

func CreatePortForwardingRulePublicFromRead(portForwardRule *PortForwardRuleRead) *PortForwardingRulePublic {
	id, name, err := portForwardRule.GetIdAndName()

	if err != nil {
		log.Infof("failed to get id from port forward rule read. details: %s", err)
		id = ""
		name = ""
	}

	externalPort, _ := strconv.Atoi(portForwardRule.Destination.Port)
	localPort, _ := strconv.Atoi(portForwardRule.LocalPort)

	return &PortForwardingRulePublic{
		ID:              id,
		Name:            name,
		ExternalAddress: net.ParseIP(portForwardRule.Destination.Address),
		ExternalPort:    externalPort,
		LocalAddress:    net.ParseIP(portForwardRule.Target),
		LocalPort:       localPort,
	}
}
