package intializer

import (
	"context"
	"encoding/json"
	"fmt"

	nvresourcebetav1 "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	applyconfigurationsresourcev1 "k8s.io/client-go/applyconfigurations/resource/v1"
)

// EnsureResourceClaimTemplates ensures all ResourceClaimTemplates exist and are up to date via SSA.
func EnsureResourceClaimTemplates() error {
	ctx := context.TODO()

	for _, zone := range config.Config.Zones {
		client := zone.K8s.Client.ResourceV1()
		ns := zone.K8s.Namespaces.Deployment

		for _, rct := range zone.ResourceClaimTemplates {
			applied, err := client.
				ResourceClaimTemplates(ns).
				Apply(ctx, CreateSharedGPUClaimTemplateSSA(models.ResourceClaimTemplatePublic{
					Name:             rct.Name,
					Namespace:        ns,
					DeviceClass:      rct.DeviceClass,
					Requests:         rct.Requests,
					Driver:           rct.Driver,
					Strategy:         rct.Strategy,
					MPSActiveThreads: rct.MPSActiveThreads,
					MPSMemoryLimit:   rct.MPSMemoryLimit,
				}), v1.ApplyOptions{
					FieldManager: "go-deploy",
					Force:        true,
				})
			if err != nil {
				// TODO: actual logging
				fmt.Printf("Failed to apply ResourceClaimTemplate %s/%s: %v\n", ns, rct.Name, err)
				continue
			}

			// TODO: actual logging
			fmt.Printf("Applied ResourceClaimTemplate (SSA): %s/%s\n", ns, applied.Name)
		}
	}

	return nil
}

func CreateSharedGPUClaimTemplateSSA(rct models.ResourceClaimTemplatePublic) *applyconfigurationsresourcev1.ResourceClaimTemplateApplyConfiguration {

	var devreqs []*applyconfigurationsresourcev1.DeviceRequestApplyConfiguration = make([]*applyconfigurationsresourcev1.DeviceRequestApplyConfiguration, 0, len(rct.Requests))

	for _, req := range rct.Requests {
		devreqs = append(devreqs, applyconfigurationsresourcev1.DeviceRequest().
			WithName(req).
			WithExactly(
				applyconfigurationsresourcev1.ExactDeviceRequest().
					WithDeviceClassName(rct.DeviceClass).WithCount(1), // TODO: make configurable
			),
		)
	}

	dfltPdevmemlim, _ := resource.ParseQuantity(rct.MPSMemoryLimit)

	params, _ := json.Marshal(&nvresourcebetav1.GpuConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "resource.nvidia.com/v1beta1",
			Kind:       "GpuConfig",
		},
		Sharing: &nvresourcebetav1.GpuSharing{
			Strategy: nvresourcebetav1.GpuSharingStrategy(rct.Strategy),
			MpsConfig: &nvresourcebetav1.MpsConfig{
				DefaultActiveThreadPercentage:  &rct.MPSActiveThreads,
				DefaultPinnedDeviceMemoryLimit: &dfltPdevmemlim,
			},
		},
	})

	arct := applyconfigurationsresourcev1.ResourceClaimTemplate(rct.Name, rct.Namespace).
		WithSpec(
			applyconfigurationsresourcev1.ResourceClaimTemplateSpec().
				WithSpec(
					applyconfigurationsresourcev1.ResourceClaimSpec().
						WithDevices(
							applyconfigurationsresourcev1.DeviceClaim().
								WithRequests(
									devreqs...,
								).WithConfig(applyconfigurationsresourcev1.DeviceClaimConfiguration().
								WithRequests(rct.Requests...).
								WithOpaque(&applyconfigurationsresourcev1.OpaqueDeviceConfigurationApplyConfiguration{
									Driver: &rct.Driver,
									Parameters: &runtime.RawExtension{
										Raw: params,
									},
								},
								),
							),
						),
				),
		)

	if arct.Labels == nil {
		arct.Labels = map[string]string{}
	}
	arct.Labels["managed-by"] = "go-deploy"

	return arct
}
