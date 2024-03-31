package body

type GpuGroupRead struct {
	Name      string `json:"name"`
	Total     int    `json:"total"`
	Available int    `json:"available"`
}
