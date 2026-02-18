package status_update

import (
	"context"
	"fmt"
	"time"

	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/k8s_service"
	"github.com/kthcloud/go-deploy/utils"
)

func GpuClaimStatusListener(ctx context.Context) error {
	for _, zone := range config.Config.EnabledZones() {
		if !zone.HasCapability(configModels.ZoneCapabilityDeployment) && !zone.HasCapability(configModels.ZoneCapabilityDRA) {
			continue
		}

		log.Println("Setting up gpu claim status watcher for zone", zone.Name)

		z := zone
		err := k8s_service.New().SetupResourceClaimWatcher(ctx, &z, func(name string, status models.ResourceClaimStatus, action string) {
			log.Println("New gpu claim event!")

			err := gpu_claim_repo.New().ReconcileStateByName(name,
				gpu_claim_repo.WithSetStatus(parseGpuClaimStatus(status)),
				gpu_claim_repo.WithSetConsumers(convertGpuClaimConsumers(status.Consumers)),
				gpu_claim_repo.WithSetAllocated(convertGpuClaimAllocations(status.AllocationResults)),
			)
			if err != nil {
				// will happen if the resourceclaim isnt in the db
				log.Println("Failed to set reconcile state for gpu claim", name, "details:", err)
				return
			}
		})
		if err != nil {
			return fmt.Errorf("failed to set up gpu claim state reconciler for zone %s. details: %w", zone.Name, err)
		}
	}
	return nil
}

func parseGpuClaimStatus(status models.ResourceClaimStatus) *model.GpuClaimStatus {
	return &model.GpuClaimStatus{
		Phase: func() model.GpuClaimStatusPhase {
			if status.Allocated {
				return model.GpuClaimStatusPhase_Bound
			}
			return model.GpuClaimStatusPhase_Pending
		}(),
		LastSynced: utils.PtrOf(time.Now()),
	}
}

func convertGpuClaimConsumers(consumers []models.ResourceClaimConsumerPublic) []model.GpuClaimConsumer {
	res := make([]model.GpuClaimConsumer, len(consumers))
	for i, consumer := range consumers {
		res[i] = model.GpuClaimConsumer{
			APIGroup: consumer.APIGroup,
			Name:     consumer.Name,
			Resource: consumer.Resource,
			UID:      consumer.UID,
		}
	}
	return res
}

func convertGpuClaimAllocations(allocations []models.ResourceClaimAllocationResultPublic) map[string][]model.AllocatedGpu {
	res := make(map[string][]model.AllocatedGpu)
	for _, alloc := range allocations {
		res[alloc.Request] = append(res[alloc.Request], model.AllocatedGpu{
			Pool:        alloc.Pool,
			Device:      alloc.Device,
			ShareID:     alloc.ShareID,
			AdminAccess: alloc.AdminAccess,
		})
	}
	return res
}
