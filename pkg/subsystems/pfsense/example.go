package pfsense

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/pfsense/models"
	"log"
	"net"
)

func Example() {
	pfSenseConf := conf.Env.PfSense

	client, _ := New(&ClientConf{
		ApiUrl:         pfSenseConf.Url,
		Username:       pfSenseConf.Identity,
		Password:       pfSenseConf.Secret,
		PublicIP:       net.ParseIP(pfSenseConf.PublicIP),
		PortRangeStart: pfSenseConf.PortRangeStart,
		PortRangeEnd:   pfSenseConf.PortRangeEnd,
	})

	name := "demo"

	public := models.PortForwardingRulePublic{
		Name:         name,
		LocalAddress: net.ParseIP("172.31.1.69"),
		LocalPort:    22,
	}

	id, err := client.CreatePortForwardingRule(&public)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("created rule with id ", id)

	rule, err := client.ReadPortForwardingRule(id)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("result: ")
	fmt.Println(rule)

	err = client.DeletePortForwardingRule(rule.ID)
	if err != nil {
		log.Fatalln(err)
	}
}
