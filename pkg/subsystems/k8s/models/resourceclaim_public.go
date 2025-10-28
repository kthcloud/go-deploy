package models

import (
	"bytes"
	"slices"
	"time"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ResourceClaimPublic struct {
	Name              string                                `bson:"name"`
	Namespace         string                                `bson:"namespace"`
	DeviceRequests    []ResourceClaimDeviceRequestPublic    `bson:"deviceRequests"`
	Allocated         bool                                  `bson:"allocated"`
	AllocationResults []ResourceClaimAllocationResultPublic `bson:"allocationResults,omitempty"`
	Consumers         []ResourceClaimConsumerPublic         `bson:"consumers,omitempty"`
	CreatedAt         time.Time                             `bson:"createdAt"`
}

type ResourceClaimDeviceRequestPublic struct {
	Name                   string                             `bson:"name"`
	RequestsExactly        *ResourceClaimExactlyRequestPublic `bson:"requestExactly,omitempty"`
	RequestsFirstAvaliable []ResourceClaimBaseRequestPublic   `bson:"requestFirstAvailable,omitempty"`
	Config                 *ResourceClaimOpaquePublic         `bson:"opaque,omitempty"`
}

type ResourceClaimBaseRequestPublic struct {
	AllocationMode   string                                         `bson:"allocationMode"`
	CapacityRequests map[resourcev1.QualifiedName]resource.Quantity `bson:"capacityRequests,omitempty"`
	Count            int64                                          `bson:"count"`
	DeviceClassName  string                                         `bson:"deviceClassName"`
	SelectorCelExprs []string                                       `bson:"selectorCelExprs,omitempty"`
	//TODO: Tolerations
}

type ResourceClaimExactlyRequestPublic struct {
	ResourceClaimBaseRequestPublic `bson:",inline"`
	AdminAccess                    *bool `bson:"adminAccess"`
}

type ResourceClaimOpaquePublic struct {
	Driver     string           `bson:"driver"`
	Parameters dra.OpaqueParams `bson:"parameters"`
}

type ResourceClaimAllocationResultPublic struct {
	Pool        string `bson:"pool,omitempty"`
	Device      string `bson:"device,omitempty"`
	Request     string `bson:"request,omitempty"`
	ShareID     string `bson:"shareID,omitempty"`
	AdminAccess bool   `bson:"adminAccess,omitempty"`
}

type ResourceClaimConsumerPublic struct {
	APIGroup string `bson:"apiGroup,omitempty"`
	Resource string `bson:"resource,omitempty"`
	Name     string `bson:"name,omitempty"`
	UID      string `bson:"uid,omitempty"`
}

func CreateResourceClaimPublicFromRead(claim *resourcev1.ResourceClaim) *ResourceClaimPublic {
	if claim == nil {
		return nil
	}

	rc := &ResourceClaimPublic{
		Name:      claim.Name,
		Namespace: claim.Namespace,
		CreatedAt: formatCreatedAt(claim.Annotations),
	}

	rc.DeviceRequests = make([]ResourceClaimDeviceRequestPublic, 0, len(claim.Spec.Devices.Requests))
	for _, req := range claim.Spec.Devices.Requests {
		var request ResourceClaimDeviceRequestPublic = ResourceClaimDeviceRequestPublic{
			Name: req.Name,
		}

		if req.Exactly != nil {
			request.RequestsExactly = &ResourceClaimExactlyRequestPublic{
				ResourceClaimBaseRequestPublic: ResourceClaimBaseRequestPublic{
					AllocationMode:   string(req.Exactly.AllocationMode),
					CapacityRequests: req.Exactly.Capacity.Requests,
					Count:            req.Exactly.Count,
					DeviceClassName:  req.Exactly.DeviceClassName,
				},
				AdminAccess: req.Exactly.AdminAccess,
			}
			request.RequestsExactly.SelectorCelExprs = make([]string, 0, len(req.Exactly.Selectors))
			for _, sel := range req.Exactly.Selectors {
				if sel.CEL != nil && sel.CEL.Expression != "" {
					request.RequestsExactly.SelectorCelExprs = append(request.RequestsExactly.SelectorCelExprs, sel.CEL.Expression)
				}
			}
		}
		if len(req.FirstAvailable) > 0 {
			request.RequestsFirstAvaliable = make([]ResourceClaimBaseRequestPublic, 0, len(req.FirstAvailable))
			for _, subReq := range req.FirstAvailable {
				br := ResourceClaimBaseRequestPublic{
					AllocationMode:   string(subReq.AllocationMode),
					CapacityRequests: subReq.Capacity.Requests,
					Count:            subReq.Count,
					DeviceClassName:  subReq.DeviceClassName,
				}
				br.SelectorCelExprs = make([]string, 0, len(req.Exactly.Selectors))
				for _, sel := range req.Exactly.Selectors {
					if sel.CEL != nil && sel.CEL.Expression != "" {
						br.SelectorCelExprs = append(br.SelectorCelExprs, sel.CEL.Expression)
					}
				}
				request.RequestsFirstAvaliable = append(request.RequestsFirstAvaliable, br)
			}
		}

		rc.DeviceRequests = append(rc.DeviceRequests, request)
	}

	for _, cfg := range claim.Spec.Devices.Config {

		if cfg.Opaque != nil && len(cfg.Opaque.Parameters.Raw) > 0 {

			opaqueParams, err := parsers.Parse[dra.OpaqueParams](bytes.NewReader(cfg.Opaque.Parameters.Raw))
			if err != nil {
				// TODO: handle/log this error, but just stkip the invalid ones for now
				continue
			}

			for i, req := range rc.DeviceRequests {
				if slices.Contains(cfg.Requests, req.Name) {
					rc.DeviceRequests[i].Config = &ResourceClaimOpaquePublic{
						Driver:     cfg.Opaque.Driver,
						Parameters: opaqueParams,
					}
				}
			}
		}

	}

	if claim.Status.Allocation != nil {
		rc.Allocated = true
		for _, res := range claim.Status.Allocation.Devices.Results {
			ar := ResourceClaimAllocationResultPublic{
				Pool:    res.Pool,
				Device:  res.Device,
				Request: res.Request,
			}
			if res.ShareID != nil {
				ar.ShareID = string(*res.ShareID)
			}
			if res.AdminAccess != nil {
				ar.AdminAccess = *res.AdminAccess
			}

			rc.AllocationResults = append(rc.AllocationResults, ar)
		}

		for _, consumer := range claim.Status.ReservedFor {
			rc.Consumers = append(rc.Consumers, ResourceClaimConsumerPublic{
				APIGroup: consumer.APIGroup,
				Resource: consumer.Resource,
				Name:     consumer.Name,
				UID:      string(consumer.UID),
			})
		}
	}

	return rc
}
