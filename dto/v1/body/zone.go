package body

type ZoneRead struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	Interface    *string  `json:"interface"`

	// Type
	// Deprecated: use capabilities instead
	Type string `json:"type"`
}
