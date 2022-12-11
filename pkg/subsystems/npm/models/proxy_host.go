package models

type ProxyHost struct {
	ID                    int      `json:"id"`
	CreatedOn             string   `json:"created_on"`
	DomainNames           []string `json:"domain_names"`
	ForwardHost           string   `json:"forward_host"`
	ForwardPort           int      `json:"forward_port"`
	CertificateID         int      `json:"certificate_id"`
	SslForced             int      `json:"ssl_forced"`
	AllowWebsocketUpgrade int      `json:"allow_websocket_upgrade"`
	ForwardScheme         string   `json:"forward_scheme"`
	Enabled               int      `json:"enabled"`
}