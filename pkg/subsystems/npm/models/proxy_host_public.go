package models

type ProxyHostPublic struct {
	ID                    int        `json:"id" bson:"id"`
	DomainNames           []string   `json:"domain_names" bson:"domain_names"`
	ForwardHost           string     `json:"forward_host" bson:"forward_host"`
	ForwardPort           int        `json:"forward_port" bson:"forward_port"`
	CertificateID         int        `json:"certificate_id" bson:"certificate_id"`
	AllowWebsocketUpgrade bool       `json:"allow_websocket_upgrade" bson:"allow_websocket_upgrade"`
	ForwardScheme         string     `json:"forward_scheme" bson:"forward_scheme"`
	Enabled               bool       `json:"enabled" bson:"enabled"`
	Locations             []Location `json:"locations" bson:"locations"`
}

func CreateProxyHostPublicFromCreatedBody(created *ProxyHostCreated) *ProxyHostPublic {
	return &ProxyHostPublic{
		ID:                    created.ID,
		DomainNames:           created.DomainNames,
		ForwardHost:           created.ForwardHost,
		ForwardPort:           created.ForwardPort,
		CertificateID:         created.CertificateID,
		AllowWebsocketUpgrade: intToBool(created.AllowWebsocketUpgrade),
		ForwardScheme:         created.ForwardScheme,
		Enabled:               intToBool(created.Enabled),
		Locations:             created.Locations,
	}
}

func CreateProxyHostPublicFromReadBody(read *ProxyHostRead) *ProxyHostPublic {
	return &ProxyHostPublic{
		ID:                    read.ID,
		DomainNames:           read.DomainNames,
		ForwardHost:           read.ForwardHost,
		ForwardPort:           read.ForwardPort,
		CertificateID:         read.CertificateID,
		AllowWebsocketUpgrade: intToBool(read.AllowWebsocketUpgrade),
		ForwardScheme:         read.ForwardScheme,
		Enabled:               intToBool(read.Enabled),
		Locations:             read.Locations,
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i == 1
}
