package models

type ProxyHostCreated struct {
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
		LetsencryptAgree bool `json:"letsencrypt_agree"`
		DNSChallenge     bool `json:"dns_challenge"`
	} `json:"meta"`
	AllowWebsocketUpgrade int        `json:"allow_websocket_upgrade"`
	HTTP2Support          int        `json:"http2_support"`
	ForwardScheme         string     `json:"forward_scheme"`
	Enabled               int        `json:"enabled"`
	Locations             []Location `json:"locations"`
	HstsEnabled           int        `json:"hsts_enabled"`
	HstsSubdomains        int        `json:"hsts_subdomains"`
	Certificate           struct {
		OwnerUserID int      `json:"owner_user_id"`
		Provider    string   `json:"provider"`
		NiceName    string   `json:"nice_name"`
		DomainNames []string `json:"domain_names"`
		ExpiresOn   string   `json:"expires_on"`
		Meta        struct {
		} `json:"meta"`
	} `json:"certificate"`
	Owner struct {
		IsDisabled int    `json:"is_disabled"`
		Name       string `json:"name"`
		Nickname   string `json:"nickname"`
		Avatar     string `json:"avatar"`
	} `json:"owner"`
	AccessList         interface{} `json:"access_list"`
	UseDefaultLocation bool        `json:"use_default_location"`
	Ipv6               bool        `json:"ipv6"`
}
