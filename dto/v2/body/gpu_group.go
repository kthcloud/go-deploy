package body

type GpuGroupRead struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Zone        string `json:"zone"`
	Vendor      string `json:"vendor"`

	Total     int `json:"total"`
	Available int `json:"available"`
}
