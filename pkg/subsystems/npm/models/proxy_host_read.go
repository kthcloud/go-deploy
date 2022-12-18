package models

type ProxyHostRead struct {
	ID             int      `json:"id"`
	CreatedOn      string   `json:"created_on"`
	ModifiedOn     string   `json:"modified_on"`
	OwnerUserID    int      `json:"owner_user_id"`
	DomainNames    []string `json:"domain_names"`
	ForwardHost    string   `json:"forward_host"`
	ForwardPort    int      `json:"forward_port"`
	AccessListID   int      `json:"access_list_id"`
	CertificateID  int      `json:"certificate_id"`
	SslForced      int      `json:"ssl_forced"`
	CachingEnabled int      `json:"caching_enabled"`
	BlockExploits  int      `json:"block_exploits"`
	AdvancedConfig string   `json:"advanced_config"`
	Meta           struct {
		NginxOnline bool        `json:"nginx_online"`
		NginxErr interface{} `json:"nginx_err"`
	} `json:"meta"`
	AllowWebsocketUpgrade int        `json:"allow_websocket_upgrade"`
	HTTP2Support          int        `json:"http2_support"`
	ForwardScheme         string     `json:"forward_scheme"`
	Enabled               int        `json:"enabled"`
	Locations             []Location `json:"locations"`
	HstsEnabled           int        `json:"hsts_enabled"`
	HstsSubdomains        int        `json:"hsts_subdomains"`
}

