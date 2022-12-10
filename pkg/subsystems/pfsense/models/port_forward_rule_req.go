package models

import (
	"net"
	"strconv"
)

// PortForwardRuleCreateReq defines parameters for APIFirewallNATOutboundPortForwardCreate.
type PortForwardRuleCreateReq struct {
	// Apply Specify whether you would like this rule to be applied immediately, or simply written to the configuration to be applied later. If you are modifying multiple rules at once it is best to set this to false and apply the changes afterwards using the `/api/v1/firewall/apply` endpoint.
	Apply bool `json:"apply,omitempty"`

	// Descr Description for the rule.
	Descr string `json:"descr,omitempty"`

	// Disabled Disable this rule.
	Disabled bool `json:"disabled,omitempty"`

	// Dst Destination address of the port forward rule. This may be a single IP, network CIDR, alias name, or interface. When specifying an interface, you may use the real interface ID (e.g. igb0), the descriptive interface name, or the pfSense ID (e.g. wan, lan, optx). To use only the  interface's assigned address, add `ip` to the end of the interface name otherwise  the entire interface's subnet is implied. To negate the context of the destination address, you may prefix the value with `!`.
	Dst string `json:"dst"`

	// Dstport TCP and/or UDP destination port, port range or port alias to apply to this rule. You may specify `any` to match any destination port. This parameter is required when `protocol` is set to `tcp`, `udp`, or `tcp/udp`.
	Dstport string `json:"dstport,omitempty"`

	// Interface Interface the rule will apply to. You may specify either the interface's descriptive name, the pfSense ID (wan, lan, optx), or the real interface ID (e.g. igb0).
	Interface string `json:"interface"`

	// LocalPort TCP and/or UDP port, port range, or port alias to forward traffic to on the `target`. This parameter is required when `protocol` is set to `tcp`, `udp`, or `tcp/udp`.
	LocalPort string `json:"local-port,omitempty"`

	// Natreflection NAT reflection mode to use for this rule. Leave unspecified to use system default.
	Natreflection string `json:"natreflection,omitempty"`

	// Nordr Disable redirection for traffic matching this rule.
	Nordr bool `json:"nordr,omitempty"`

	// Nosync Prevent this rule from automatically syncing to other CARP members.
	Nosync bool `json:"nosync,omitempty"`

	// Protocol Transfer protocol this rule will apply to.
	Protocol string `json:"protocol"`

	// Src Source address of the port forward rule. This may be a single IP, network CIDR, alias name, or interface. When specifying an interface, you may use the real interface ID (e.g. igb0), the descriptive interface name, or the pfSense ID (e.g. wan, lan, optx). To use only the  interface's assigned address, add `ip` to the end of the interface name otherwise  the entire interface's subnet is implied. To negate the context of the source address, you may prefix the value with `!`.
	Src string `json:"src"`

	// Srcport TCP and/or UDP source port, port range or port alias to apply to this rule. You may specify `any` to match any source port. This parameter is required when `protocol` is set to `tcp`, `udp`, or `tcp/udp`.
	Srcport string `json:"srcport,omitempty"`

	// Target IP address or alias to forward traffic to.
	Target string `json:"target"`

	// Top Place this firewall rule at the top of the access control list.
	Top bool `json:"top,omitempty"`
}

func CreatePortForwardRuleCreateReq(externalIP net.IP, externalPort int, internalIP net.IP, internalPort int, description string) PortForwardRuleCreateReq {
	reqBody := PortForwardRuleCreateReq{
		Apply:         true,
		Descr:         description,
		Disabled:      false,
		Dst:           externalIP.String(),
		Dstport:       strconv.Itoa(externalPort),
		Interface:     "WAN",
		LocalPort:     strconv.Itoa(internalPort),
		Natreflection: "enable",
		Nordr:         false,
		Nosync:        false,
		Protocol:      "tcp",
		Src:           "any",
		Srcport:       "any",
		Target:        internalIP.String(),
		Top:           false,
	}
	return reqBody
}
