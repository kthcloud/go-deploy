package v2

import (
	"context"
	"testing"

	"github.com/kthcloud/go-deploy/models/model"
	"go.mongodb.org/mongo-driver/bson"
)

func TestExtractIDAndParams(t *testing.T) {
	t.Run("nil job", func(t *testing.T) {
		if _, _, err := extractIDAndParams(nil); err == nil {
			t.Fatal("expected error for nil job")
		}
	})

	t.Run("non-string id does not panic", func(t *testing.T) {
		job := &model.Job{Args: map[string]any{"id": 123, "params": map[string]any{}}}
		if _, _, err := extractIDAndParams(job); err == nil {
			t.Fatal("expected error for non-string id")
		}
	})

	t.Run("empty id", func(t *testing.T) {
		job := &model.Job{Args: map[string]any{"id": "", "params": map[string]any{}}}
		if _, _, err := extractIDAndParams(job); err == nil {
			t.Fatal("expected error for empty id")
		}
	})

	t.Run("nil params", func(t *testing.T) {
		job := &model.Job{Args: map[string]any{"id": "claim-1", "params": map[string]any(nil)}}
		if _, _, err := extractIDAndParams(job); err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("accepts bson.M params", func(t *testing.T) {
		job := &model.Job{Args: map[string]any{"id": "claim-1", "params": bson.M{"name": "test"}}}
		id, params, err := extractIDAndParams(job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "claim-1" || params["name"] != "test" {
			t.Fatalf("unexpected result: id=%q params=%v", id, params)
		}
	})
}

func TestDecodeGpuClaimParams(t *testing.T) {
	t.Run("create nil raw", func(t *testing.T) {
		if _, err := DecodeGpuClaimCreateParams(nil); err == nil {
			t.Fatal("expected error for nil raw")
		}
	})

	t.Run("update nil raw", func(t *testing.T) {
		if _, err := DecodeGpuClaimUpdateParams(nil); err == nil {
			t.Fatal("expected error for nil raw")
		}
	})

	t.Run("create with nil config parameters", func(t *testing.T) {
		count := int64(2)
		params, err := DecodeGpuClaimCreateParams(map[string]any{
			"name":         "claim-a",
			"zone":         "local",
			"allowedRoles": []string{"admin"},
			"requested": []map[string]any{
				{
					"name":            "gpu-0",
					"allocationMode":  "ExactCount",
					"deviceClassName": "gpu.nvidia.com",
					"capacity":        map[string]string{"memory": "16Gi"},
					"count":           count,
					"selectors":       []string{"device.driver == 'gpu.nvidia.com'"},
					"config": map[string]any{
						"driver":     "gpu.nvidia.com",
						"parameters": nil,
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.Name != "claim-a" || params.Zone != "local" {
			t.Fatalf("unexpected name/zone: %+v", params)
		}
		if len(params.AllowedRoles) != 1 || params.AllowedRoles[0] != "admin" {
			t.Fatalf("unexpected allowedRoles: %+v", params.AllowedRoles)
		}
		if len(params.Requested) != 1 {
			t.Fatalf("expected 1 requested gpu, got %d", len(params.Requested))
		}
		req := params.Requested[0]
		if req.Capacity["memory"] != "16Gi" {
			t.Fatalf("unexpected capacity: %v", req.Capacity)
		}
		if req.Count == nil || *req.Count != count {
			t.Fatalf("unexpected count: %v", req.Count)
		}
		if len(req.Selectors) != 1 {
			t.Fatalf("unexpected selectors: %v", req.Selectors)
		}
		if req.Config == nil || req.Config.Driver != "gpu.nvidia.com" || req.Config.Parameters != nil {
			t.Fatalf("unexpected config: %+v", req.Config)
		}
	})

	t.Run("update partial without requested", func(t *testing.T) {
		params, err := DecodeGpuClaimUpdateParams(map[string]any{
			"name": "renamed",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.Name == nil || *params.Name != "renamed" {
			t.Fatalf("unexpected name: %v", params.Name)
		}
		if params.Requested != nil {
			t.Fatalf("expected nil requested, got %v", params.Requested)
		}
	})
}

func TestGpuClaimJobsInvalidArgsNoPanic(t *testing.T) {
	ctx := context.Background()
	badJob := &model.Job{Args: map[string]any{"id": 42}}

	if err := CreateGpuClaim(ctx, badJob); err == nil {
		t.Fatal("expected error from CreateGpuClaim")
	}
	if err := DeleteGpuClaim(ctx, badJob); err == nil {
		t.Fatal("expected error from DeleteGpuClaim")
	}
	if err := UpdateGpuClaim(ctx, badJob); err == nil {
		t.Fatal("expected error from UpdateGpuClaim")
	}
}
