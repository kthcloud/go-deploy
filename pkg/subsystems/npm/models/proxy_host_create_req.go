package models

type ProxyHostCreateReq struct {
	DomainNames           []string `json:"domain_names,omitempty"`
	ForwardHost           string   `json:"forward_host,omitempty"`
	ForwardPort           int      `json:"forward_port,omitempty"`
	AccessListID          int      `json:"access_list_id,omitempty"`
	CertificateID         int      `json:"certificate_id,omitempty"`
	SslForced             int      `json:"ssl_forced,omitempty"`
	CachingEnabled        int      `json:"caching_enabled,omitempty"`
	BlockExploits         int      `json:"block_exploits,omitempty"`
	AdvancedConfig        string   `json:"advanced_config,omitempty"`
	AllowWebsocketUpgrade int      `json:"allow_websocket_upgrade,omitempty"`
	HTTP2Support          int      `json:"http2_support,omitempty"`
	ForwardScheme         string   `json:"forward_scheme,omitempty"`
	Enabled               int      `json:"enabled,omitempty"`
	Locations             []string `json:"locations,omitempty"`
	HstsEnabled           int      `json:"hsts_enabled,omitempty"`
	HstsSubdomains        int      `json:"hsts_subdomains,omitempty"`
}

func CreateProxyHostBody(domainNames []string, forwardHost string, port int, certificateID int) ProxyHostCreateReq {
	return ProxyHostCreateReq{
		DomainNames:           domainNames,
		ForwardHost:           forwardHost,
		ForwardPort:           port,
		AccessListID:          0,
		CertificateID:         certificateID,
		SslForced:             1,
		CachingEnabled:        0,
		BlockExploits:         0,
		AdvancedConfig:        "",
		AllowWebsocketUpgrade: 1,
		HTTP2Support:          0,
		ForwardScheme:         "http",
		Enabled:               1,
		Locations:             nil,
		HstsEnabled:           0,
		HstsSubdomains:        0,
	}
}
