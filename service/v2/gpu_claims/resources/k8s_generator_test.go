package resources

import (
	"reflect"
	"testing"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/utils"
	v1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceClaims(t *testing.T) {
	count := int64(2)
	gc := &model.GpuClaim{
		Name: "gpu-claim",
		Requested: map[string]model.RequestedGpu{
			"gpu1": {

				AllocationMode:  "All",
				DeviceClassName: "nvidia-tesla",
				Count:           &count,
				Capacity: map[string]string{
					"memory": "16Gi",
					"cores":  "8",
					"bad":    "not-a-qty", // should be skipped
				},
				Selectors: []string{"zone == us-central1"},
			},
		},
	}

	kg := &K8sGenerator{
		gc:        gc,
		namespace: "default",
	}

	got := kg.ResourceClaims()
	if len(got) != 1 {
		t.Fatalf("expected 1 ResourceClaimPublic, got %d", len(got))
	}

	rc := got[0]

	if rc.Name != "gpu-claim" {
		t.Errorf("expected Name=gpu-claim, got %s", rc.Name)
	}
	if rc.Namespace != "default" {
		t.Errorf("expected Namespace=default, got %s", rc.Namespace)
	}
	if len(rc.DeviceRequests) != 1 {
		t.Fatalf("expected 1 DeviceRequest, got %d", len(rc.DeviceRequests))
	}

	dr := rc.DeviceRequests[0]
	if dr.Name != "gpu1" {
		t.Errorf("expected DeviceRequest.Name=gpu1, got %s", dr.Name)
	}

	exact := dr.RequestsExactly
	if exact == nil {
		t.Fatalf("expected RequestsExactly to be non-nil")
	}
	if exact.AllocationMode != "All" {
		t.Errorf("expected AllocationMode=All, got %s", exact.AllocationMode)
	}
	if exact.DeviceClassName != "nvidia-tesla" {
		t.Errorf("expected DeviceClassName=nvidia-tesla, got %s", exact.DeviceClassName)
	}
	if exact.Count != utils.ZeroDeref(&count) {
		t.Errorf("expected Count=%d, got %d", count, exact.Count)
	}

	// Validate quantities
	expectedQty, _ := resource.ParseQuantity("16Gi")
	if !reflect.DeepEqual(exact.CapacityRequests[v1.QualifiedName("memory")], expectedQty) {
		t.Errorf("expected CapacityRequests[memory]=%v, got %v",
			expectedQty, exact.CapacityRequests["memory"])
	}

	// Ensure bad quantity is skipped
	if _, ok := exact.CapacityRequests[v1.QualifiedName("bad")]; ok {
		t.Errorf("expected bad capacity key to be skipped")
	}

}
