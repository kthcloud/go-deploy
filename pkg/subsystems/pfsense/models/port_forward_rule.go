package models

type PortForwardRule struct {
	Source struct {
		Any string `json:"any"`
	} `json:"source"`
	Destination struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	} `json:"destination"`
	Ipprotocol       string `json:"ipprotocol"`
	Protocol         string `json:"protocol"`
	Target           string `json:"target"`
	LocalPort        string `json:"local-port"`
	Interface        string `json:"interface"`
	Descr            string `json:"descr"`
	AssociatedRuleID string `json:"associated-rule-id"`
	Created          struct {
		Time     string `json:"time"`
		Username string `json:"username"`
	} `json:"created"`
	Updated struct {
		Time     string `json:"time"`
		Username string `json:"username"`
	} `json:"updated"`
}
