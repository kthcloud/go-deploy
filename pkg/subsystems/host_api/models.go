package host_api

type NodeInfo struct {
	Zone string `bson:"zone" json:"zone"`
}

type GPU struct {
	Name     string `bson:"name" json:"name"`
	Slot     string `bson:"slot" json:"slot"`
	Vendor   string `bson:"vendor" json:"vendor"`
	VendorID string `bson:"vendorId" json:"vendorId"`
	Bus      string `bson:"bus" json:"bus"`
	DeviceID string `bson:"deviceId" json:"deviceId"`
	Zone     string `bson:"zone" json:"zone"`
}

type Capacities struct {
	CPU struct {
		Cores  int    `json:"cores" bson:"cores"`
		Vendor string `json:"vendor" bson:"vendor"`
	} `json:"cpu" bson:"cpu"`
	RAM struct {
		Total int `json:"total" bson:"total"`
	} `json:"ram" bson:"ram"`
	GPU struct {
		Count int `json:"count" bson:"count"`
	} `json:"gpu" bson:"gpu"`
}

type Status struct {
	CPU struct {
		Temp struct {
			Main  float64 `json:"main" bson:"main"`
			Cores []int   `json:"cores" bson:"cores"`
			Max   float64 `json:"max" bson:"max"`
		} `json:"temp" bson:"temp"`
		Load struct {
			Main  float64 `json:"main" bson:"main"`
			Cores []int   `json:"cores" bson:"cores"`
			Max   float64 `json:"max" bson:"max"`
		} `json:"load" bson:"load"`
	} `json:"cpu" bson:"cpu"`
	RAM struct {
		Load struct {
			Main float64 `json:"main" bson:"main"`
		} `json:"load" bson:"load"`
	} `json:"ram" bson:"ram"`

	GPU *struct {
		Temp []struct {
			Main float64 `json:"main" bson:"main"`
		} `json:"temp" bson:"temp"`
	} `json:"gpu,omitempty" bson:"gpu,omitempty"`
}

type GpuInfo struct {
	Name        string `bson:"name" json:"name"`
	Slot        string `bson:"slot" json:"slot"`
	Vendor      string `bson:"vendor" json:"vendor"`
	VendorID    string `bson:"vendorId" json:"vendorId"`
	Bus         string `bson:"bus" json:"bus"`
	DeviceID    string `bson:"deviceId" json:"deviceId"`
	Passthrough bool   `bson:"passthrough" json:"passthrough"`
}
