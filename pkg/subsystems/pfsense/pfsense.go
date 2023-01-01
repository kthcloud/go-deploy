package pfsense

import "net"

type Client struct {
	apiUrl         string
	username       string
	password       string
	publicIP       net.IP
	portRangeStart int
	portRangeEnd   int
}

type ClientConf struct {
	ApiUrl         string
	Username       string
	Password       string
	PublicIP       net.IP
	PortRangeStart int
	PortRangeEnd   int
}

func New(config *ClientConf) (*Client, error) {
	client := Client{
		apiUrl:         config.ApiUrl,
		username:       config.Username,
		password:       config.Password,
		publicIP:       config.PublicIP,
		portRangeStart: config.PortRangeStart,
		portRangeEnd:   config.PortRangeEnd,
	}

	return &client, nil
}
