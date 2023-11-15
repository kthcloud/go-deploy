package body

type DiscoverRead struct {
	Version string `json:"version"`
	Roles   []Role `json:"roles"`
}
