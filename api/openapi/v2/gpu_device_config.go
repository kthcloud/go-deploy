package v2

import (
	"encoding/json"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/api/nvidia"
	"go.mongodb.org/mongo-driver/bson"
)

type GpuDeviceConfigurationWrapper struct {
	GpuDeviceConfiguration `mapstructure:"-" json:"-" bson:"-"`
}

func (w *GpuDeviceConfigurationWrapper) UnmarshalJSON(data []byte) error {
	// Peek at the "driver" field first
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var driver string
	if d, ok := raw["driver"]; ok {
		if err := json.Unmarshal(d, &driver); err != nil {
			return err
		}
	}

	switch driver {
	case "gpu.nvidia.com":
		var n NvidiaDeviceConfiguration
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		if n.Parameters != nil {
			if n.Parameters.APIVersion == "" {
				n.Parameters.APIVersion = "resource.nvidia.com/v1beta1"
			}
			if n.Parameters.Kind == "" {
				n.Parameters.Kind = "GpuConfig"
			}
		}
		w.GpuDeviceConfiguration = n

	default:
		var g GenericDeviceConfiguration
		if err := json.Unmarshal(data, &g); err != nil {
			return err
		}
		w.GpuDeviceConfiguration = g
	}

	return nil
}

func (w GpuDeviceConfigurationWrapper) MarshalJSON() ([]byte, error) {
	if w.GpuDeviceConfiguration == nil {
		return []byte("null"), nil
	}
	return json.Marshal(w.GpuDeviceConfiguration)
}

// MarshalBSON implements bson.Marshaler
func (w GpuDeviceConfigurationWrapper) MarshalBSON() ([]byte, error) {
	j, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := bson.UnmarshalExtJSON(j, false, &doc); err != nil {
		return nil, err
	}
	return bson.Marshal(doc)
}

// UnmarshalBSON implements bson.Unmarshaler
func (w *GpuDeviceConfigurationWrapper) UnmarshalBSON(data []byte) error {
	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Detect driver
	driver := ""
	if d, ok := raw["driver"].(string); ok {
		driver = d
	}

	switch driver {
	case "gpu.nvidia.com":
		var n NvidiaDeviceConfiguration
		if err := bson.Unmarshal(data, &n); err != nil {
			return err
		}
		if n.Parameters != nil {
			if n.Parameters.APIVersion == "" {
				n.Parameters.APIVersion = "resource.nvidia.com/v1beta1"
			}
			if n.Parameters.Kind == "" {
				n.Parameters.Kind = "GpuConfig"
			}
		}
		w.GpuDeviceConfiguration = n
	default:
		var g GenericDeviceConfiguration
		if err := bson.Unmarshal(data, &g); err != nil {
			return err
		}
		w.GpuDeviceConfiguration = g
	}

	return nil
}

// GpuDeviceConfiguration represents a vendor-specific GPU configuration.
type GpuDeviceConfiguration interface {
	DriverName() string
	json.Marshaler
}

// GenericDeviceConfiguration is a catch-all configuration when no vendor-specific struct is used.
type GenericDeviceConfiguration struct {
	Driver string `json:"driver" bson:"driver"`
}

func (g GenericDeviceConfiguration) DriverName() string {
	return g.Driver
}

func (g GenericDeviceConfiguration) MarshalJSON() ([]byte, error) {
	type genericAlias GenericDeviceConfiguration
	return json.Marshal(&struct {
		Type string `json:"type" bson:"type"`
		genericAlias
	}{
		Type:         "generic",
		genericAlias: (genericAlias)(g),
	})
}

// NvidiaDeviceConfiguration represents NVIDIA-specific configuration options.
type NvidiaDeviceConfiguration struct {
	Driver     string            `json:"driver" bson:"driver"`
	Parameters *nvidia.GpuConfig `json:"parameters,omitempty" bson:"parameters,omitempty"`
}

func (NvidiaDeviceConfiguration) DriverName() string {
	return "gpu.nvidia.com"
}

func (n NvidiaDeviceConfiguration) MarshalJSON() ([]byte, error) {
	type nvidiaAlias NvidiaDeviceConfiguration
	return json.Marshal(&struct {
		Type string `json:"type" bson:"type"`
		nvidiaAlias
	}{
		Type:        "nvidia",
		nvidiaAlias: (nvidiaAlias)(n),
	})
}
