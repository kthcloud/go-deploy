package body

type PortRead struct {
	Name         string         `json:"name,omitempty" bson:"name"`
	Port         int            `json:"port,omitempty" bson:"port"`
	ExternalPort *int           `json:"externalPort,omitempty" bson:"externalPort,omitempty"`
	Protocol     string         `json:"protocol,omitempty" bson:"protocol,"`
	HttpProxy    *HttpProxyRead `json:"httpProxy,omitempty" bson:"httpProxy,omitempty"`
}

type PortCreate struct {
	Name      string           `json:"name" bson:"name" binding:"required,min=1,max=100"`
	Port      int              `json:"port" bson:"port" binding:"required,min=1,max=65535"`
	Protocol  string           `json:"protocol" bson:"protocol," binding:"required,oneof=tcp udp"`
	HttpProxy *HttpProxyCreate `json:"httpProxy,omitempty" bson:"httpProxy,omitempty" binding:"omitempty,dive"`
}

type PortUpdate struct {
	Name      string           `json:"name,omitempty" bson:"name" binding:"required,min=1,max=100"`
	Port      int              `json:"port,omitempty" bson:"port" binding:"required,min=1,max=65535"`
	Protocol  string           `json:"protocol,omitempty" bson:"protocol," binding:"required,oneof=tcp udp"`
	HttpProxy *HttpProxyUpdate `json:"httpProxy,omitempty" bson:"httpProxy,omitempty" binding:"omitempty"`
}

type CustomDomainRead struct {
	Domain string `json:"domain"`
	Url    string `json:"url"`
	Status string `json:"status"`
	Secret string `json:"secret"`
}

type HttpProxyRead struct {
	Name         string            `json:"name" bson:"name,omitempty" binding:"required,rfc1035,min=3,max=30"`
	URL          *string           `json:"url,omitempty,omitempty"`
	CustomDomain *CustomDomainRead `json:"customDomain,omitempty" bson:"customDomain,omitempty"`
}

type HttpProxyCreate struct {
	Name         string  `json:"name" bson:"name,omitempty" binding:"required,rfc1035,min=3,max=30"`
	CustomDomain *string `json:"customDomain,omitempty" bson:"customDomain,omitempty" binding:"omitempty,custom_domain,domain_name,min=1,max=100"`
}

type HttpProxyUpdate struct {
	Name         string  `json:"name,omitempty" bson:"name,omitempty" binding:"required,rfc1035,min=3,max=30"`
	CustomDomain *string `json:"customDomain,omitempty" bson:"customDomain,omitempty" binding:"omitempty,custom_domain,domain_name,max=100"`
}
