package body

type ZoneEndpoints struct {
	Deployment string `json:"deployment,omitempty"`
	Storage    string `json:"storage,omitempty"`
	VM         string `json:"vm,omitempty"`
	VmApp      string `json:"vmApp,omitempty"`
}

type ZoneRead struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Capabilities []string      `json:"capabilities"`
	Endpoints    ZoneEndpoints `json:"endpoints"`
	Legacy       bool          `json:"legacy"`
	Enabled      bool          `json:"enabled"`
}
