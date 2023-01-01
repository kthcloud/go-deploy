package models

type UpdateMeta struct {
	LetsencryptAgree bool `json:"letsencrypt_agree"`
	DNSChallenge     bool `json:"dns_challenge"`
}

type ProxyHostUpdate struct {
	ForwardScheme         string     `json:"forward_scheme"`
	ForwardHost           string     `json:"forward_host"`
	ForwardPort           int        `json:"forward_port"`
	AdvancedConfig        string     `json:"advanced_config"`
	DomainNames           []string   `json:"domain_names"`
	AccessListID          int        `json:"access_list_id"`
	CertificateID         int        `json:"certificate_id"`
	SslForced             bool       `json:"ssl_forced"`
	Meta                  UpdateMeta `json:"meta"`
	Locations             []Location `json:"locations"`
	BlockExploits         bool       `json:"block_exploits"`
	CachingEnabled        bool       `json:"caching_enabled"`
	AllowWebsocketUpgrade bool       `json:"allow_websocket_upgrade"`
	HTTP2Support          bool       `json:"http2_support"`
	HstsEnabled           bool       `json:"hsts_enabled"`
	HstsSubdomains        bool       `json:"hsts_subdomains"`
	Enabled               bool       `json:"enabled"`
}

func CreateProxyHostUpdateBody(public *ProxyHostPublic) ProxyHostUpdate {
	return ProxyHostUpdate{
		ForwardScheme:         public.ForwardScheme,
		ForwardHost:           public.ForwardHost,
		ForwardPort:           public.ForwardPort,
		AdvancedConfig:        "",
		DomainNames:           public.DomainNames,
		AccessListID:          0,
		CertificateID:         public.CertificateID,
		SslForced:             true,
		Meta:                  UpdateMeta{},
		Locations:             public.Locations,
		BlockExploits:         false,
		CachingEnabled:        false,
		AllowWebsocketUpgrade: public.AllowWebsocketUpgrade,
		HTTP2Support:          false,
		HstsEnabled:           false,
		HstsSubdomains:        false,
		Enabled:               public.Enabled,
	}
}
