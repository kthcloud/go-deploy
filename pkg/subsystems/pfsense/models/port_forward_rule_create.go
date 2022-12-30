package models

import (
	"strconv"
)

// PortForwardRuleCreate defines parameters for APIFirewallNATOutboundPortForwardCreate.
type PortForwardRuleCreate struct {
	// Apply Specify whether you would like this rule to be applied immediately, or simply written to the configuration to be applied later. If you are modifying multiple rules at once it is best to set this to false and apply the changes afterwards using the `/api/v1/firewall/apply` endpoint.
	Apply bool `json:"apply,omitempty"`

	// Description The description for the rule.
	Description string `json:"descr,omitempty"`

	// Disabled Disable this rule.
	Disabled bool `json:"disabled,omitempty"`

	// Destination The destination address of the port forward rule. This may be a single IP, network CIDR, alias name, or interface. When specifying an interface, you may use the real interface ID (e.g. igb0), the descriptive interface name, or the pfSense ID (e.g. wan, lan, optx). To use only the  interface's assigned address, add `ip` to the end of the interface name otherwise  the entire interface's subnet is implied. To negate the context of the destination address, you may prefix the value with `!`.
	Destination string `json:"dst"`

	// DestinationPort TCP and/or UDP destination port, port range or port alias to apply to this rule. You may specify `any` to match any destination port. This parameter is required when `protocol` is set to `tcp`, `udp`, or `tcp/udp`.
	DestinationPort string `json:"dstport,omitempty"`

	// Interface The rule will apply to. You may specify either the interface's descriptive name, the pfSense ID (wan, lan, optx), or the real interface ID (e.g. igb0).
	Interface string `json:"interface"`

	// LocalPort TCP and/or UDP port, port range, or port alias to forward traffic to on the `target`. This parameter is required when `protocol` is set to `tcp`, `udp`, or `tcp/udp`.
	LocalPort string `json:"local-port,omitempty"`

	// NatReflection NAT reflection mode to use for this rule. Leave unspecified to use system default.
	NatReflection string `json:"natreflection,omitempty"`

	// NoRdr Disable redirection for traffic matching this rule.
	NoRdr bool `json:"nordr,omitempty"`

	// NoSync Prevent this rule from automatically syncing to other CARP members.
	NoSync bool `json:"nosync,omitempty"`

	// Protocol Transfer protocol this rule will apply to.
	Protocol string `json:"protocol"`

	// Src Source address of the port forward rule. This may be a single IP, network CIDR, alias name, or interface. When specifying an interface, you may use the real interface ID (e.g. igb0), the descriptive interface name, or the pfSense ID (e.g. wan, lan, optx). To use only the  interface's assigned address, add `ip` to the end of the interface name otherwise  the entire interface's subnet is implied. To negate the context of the source address, you may prefix the value with `!`.
	Src string `json:"src"`

	// SourcePort TCP and/or UDP source port, port range or port alias to apply to this rule. You may specify `any` to match any source port. This parameter is required when `protocol` is set to `tcp`, `udp`, or `tcp/udp`.
	SourcePort string `json:"srcport,omitempty"`

	// Target IP address or alias to forward traffic to.
	Target string `json:"target"`

	// Top Place this firewall rule at the top of the access control list.
	Top bool `json:"top,omitempty"`
}

func CreatePortForwardRuleCreateBody(public *PortForwardingRulePublic) PortForwardRuleCreate {
	description := public.GetDescription()

	reqBody := PortForwardRuleCreate{
		Apply:           true,
		Description:     description,
		Disabled:        false,
		Destination:     public.ExternalAddress.String(),
		DestinationPort: strconv.Itoa(public.ExternalPort),
		Target:          public.LocalAddress.String(),
		LocalPort:       strconv.Itoa(public.LocalPort),
		Interface:       "WAN",
		NatReflection:   "enable",
		NoRdr:           false,
		NoSync:          false,
		Protocol:        "tcp",
		Src:             "any",
		SourcePort:      "any",
		Top:             false,
	}
	return reqBody
}
