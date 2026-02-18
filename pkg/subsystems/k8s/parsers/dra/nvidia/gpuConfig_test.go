package nvidia_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra/nvidia"
)

func TestGPUConfigParser(t *testing.T) {
	data := []byte(`{
		"apiVersion":"resource.nvidia.com/v1beta1",
		"kind":"GpuConfig",
		"sharing":{}
	}`)

	parser := nvidia.GPUConfigParametersParserImpl{}

	if !parser.CanParse(bytes.NewReader(data)) {
		t.Fatal("CanParse returned false for valid GPUConfig JSON")
	}

	out, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !strings.HasPrefix(out.MetaAPIVersion(), "resource.nvidia.com/") {
		t.Errorf("unexpected APIVersion: got %s", out.MetaAPIVersion())
	}
	if out.MetaKind() != "GpuConfig" {
		t.Errorf("unexpected Kind: got %s", out.MetaKind())
	}
}

func TestGPUConfigParser2(t *testing.T) {
	data := []byte(`{"apiVersion":"resource.nvidia.com/v1beta1","kind":"GpuConfig","sharing":{"mpsConfig":{"defaultActiveThreadPercentage":50,"defaultPinnedDeviceMemoryLimit":"10Gi"},"strategy":"MPS"}}`)

	parser := nvidia.GPUConfigParametersParserImpl{}

	if !parser.CanParse(bytes.NewReader(data)) {
		t.Fatal("CanParse returned false for valid GPUConfig JSON")
	}

	out, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !strings.HasPrefix(out.MetaAPIVersion(), "resource.nvidia.com/") {
		t.Errorf("unexpected APIVersion: got %s", out.MetaAPIVersion())
	}
	if out.MetaKind() != "GpuConfig" {
		t.Errorf("unexpected Kind: got %s", out.MetaKind())
	}
}
