package pfsense

import (
	"go-deploy/pkg/conf"
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

	port, err := client.CreatePortForwardingRule(name, net.ParseIP("172.31.1.69"), 22)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(port)

	rules, err := client.GetPortForwardingRules()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(rules)

	err = client.DeletePortForwardingRule(name)
	if err != nil {
		log.Fatalln(err)
	}

}
