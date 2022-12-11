package models

type Certificate struct {
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