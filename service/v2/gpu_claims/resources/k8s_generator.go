package resources

import (
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/subsystems"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"github.com/kthcloud/go-deploy/service/generators"
	"github.com/kthcloud/go-deploy/utils"
	v1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type K8sGenerator struct {
	generators.K8sGeneratorBase

	image     *string
	namespace string
	client    *k8s.Client

	gc   *model.GpuClaim
	zone *configModels.Zone
}

func K8s(gc *model.GpuClaim, zone *configModels.Zone, client *k8s.Client, namespace string) *K8sGenerator {
	kg := &K8sGenerator{
		namespace: namespace,
		client:    client,
		gc:        gc,
		zone:      zone,
	}

	return kg
}

func (kg *K8sGenerator) Namespace() *models.NamespacePublic {
	ns := models.NamespacePublic{
		Name: kg.namespace,
	}

	if n := &kg.gc.Subsystems.K8s.Namespace; subsystems.Created(n) {
		ns.CreatedAt = n.CreatedAt
	}

	return &ns
}

func (kg *K8sGenerator) ResourceClaims() []models.ResourceClaimPublic {
	if kg.gc == nil {
		return nil
	}

	var rc models.ResourceClaimPublic = models.ResourceClaimPublic{
		Name:      kg.gc.Name,
		Namespace: kg.namespace,
	}

	for name, req := range kg.gc.Requested {
		rcdr := models.ResourceClaimDeviceRequestPublic{
			Name: name,
			// TODO: add request first available in the future if it is wanted
		}

		rcer := &models.ResourceClaimExactlyRequestPublic{
			ResourceClaimBaseRequestPublic: models.ResourceClaimBaseRequestPublic{
				AllocationMode:   string(req.AllocationMode),
				Count:            utils.ZeroDeref(req.Count),
				DeviceClassName:  req.DeviceClassName,
				SelectorCelExprs: req.Selectors,
			},
		}
		rcer.CapacityRequests = make(map[v1.QualifiedName]resource.Quantity, len(req.Capacity))
		for k, v := range req.Capacity {
			qty, err := resource.ParseQuantity(v)
			if err != nil {
				// skip bad format
				continue
			}
			rcer.CapacityRequests[v1.QualifiedName(k)] = qty
		}

		rcdr.RequestsExactly = rcer
		if req.Config != nil {
			rcc := &models.ResourceClaimOpaquePublic{
				Driver: req.Config.Driver,
			}
			if req.Config.Parameters != nil {
				rcc.Parameters = req.Config.Parameters
			}
			rcdr.Config = rcc
		}

		rc.DeviceRequests = append(rc.DeviceRequests, rcdr)
	}

	return []models.ResourceClaimPublic{rc}
}
