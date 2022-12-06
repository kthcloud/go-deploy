package npm

import (
	"deploy-api-go/pkg/conf"
	"deploy-api-go/utils/subsystemutils"
	"fmt"
)

type createProxyHostRequestBody struct {
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

type createProxyHostResponseBody struct {
	ID            int      `json:"id,omitempty"`
	CreatedOn     string   `json:"created_on,omitempty"`
	DomainNames   []string `json:"domain_names,omitempty"`
	ForwardHost   string   `json:"forward_host,omitempty"`
	ForwardPort   int      `json:"forward_port,omitempty"`
	CertificateID int      `json:"certificate_id,omitempty"`
	SslForced     int      `json:"ssl_forced,omitempty"`
	ForwardScheme string   `json:"forward_scheme,omitempty"`
	Enabled       int      `json:"enabled,omitempty"`
}

type listProxyHostResponseBody struct {
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

type listProxyHostsResponseBody []listProxyHostResponseBody

type npmApiError struct {
	Error struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

type tokenRequestBody struct {
	Identity string `json:"identity,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

type tokenBody struct {
	Token string `json:"token"`
}

type certificatesBody []struct {
	ID          int      `json:"id,omitempty"`
	CreatedOn   string   `json:"created_on,omitempty"`
	ModifiedOn  string   `json:"modified_on,omitempty"`
	OwnerUserID int      `json:"owner_user_id,omitempty"`
	Provider    string   `json:"provider,omitempty"`
	NiceName    string   `json:"nice_name,omitempty"`
	DomainNames []string `json:"domain_names,omitempty"`
	ExpiresOn   string   `json:"expires_on,omitempty"`
	Meta        struct {
		LetsencryptEmail       string `json:"letsencrypt_email,omitempty"`
		DNSChallenge           bool   `json:"dns_challenge,omitempty"`
		DNSProvider            string `json:"dns_provider,omitempty"`
		DNSProviderCredentials string `json:"dns_provider_credentials,omitempty"`
		LetsencryptAgree       bool   `json:"letsencrypt_agree,omitempty"`
	} `json:"meta,omitempty"`
}

func createProxyHostBody(name string, certificateID int) createProxyHostRequestBody {
	forwardHost := fmt.Sprintf("%s.%s.svc.cluster.local", name, subsystemutils.GetPrefixedName(name))
	domainNames := []string{getFqdn(name)}
	port := conf.Env.AppPort

	return createProxyHostRequestBody{
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
