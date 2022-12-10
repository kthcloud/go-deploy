package models

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/pfsense"
	"log"
	"net"
)

func example() {
	pfsenseConf := conf.Env.PfSense

	client, _ := pfsense.New(&pfsense.ClientConf{
		ApiUrl:         pfsenseConf.Url,
		Username:       pfsenseConf.Identity,
		Password:       pfsenseConf.Secret,
		PublicIP:       net.ParseIP(pfsenseConf.PublicIP),
		PortRangeStart: pfsenseConf.PortRangeStart,
		PortRangeEnd:   pfsenseConf.PortRangeEnd,
	})
	port, err := client.CreatePortForwardRule(net.ParseIP("172.31.1.69"), 22, "deploy test")

	rules, err := client.GetPortForwardRules()

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rules)
	log.Println(port)
}
