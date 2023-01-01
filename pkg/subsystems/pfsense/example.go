package pfsense

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/pfsense/models"
	"log"
	"net"
)

func Example() {
	pfsenseConf := conf.Env.PfSense

	client, _ := New(&ClientConf{
		ApiUrl:         pfsenseConf.Url,
		Username:       pfsenseConf.Identity,
		Password:       pfsenseConf.Secret,
		PublicIP:       net.ParseIP(pfsenseConf.PublicIP),
		PortRangeStart: pfsenseConf.PortRangeStart,
		PortRangeEnd:   pfsenseConf.PortRangeEnd,
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
