package vm_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/pkg/db"
	rErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create creates a new VM.
func (client *Client) Create(id, owner string, params *model.VmCreateParams) (*model.VM, error) {
	portMap := make(map[string]model.Port)
	for _, paramPort := range params.PortMap {
		port := model.Port{
			Name:     paramPort.Name,
			Port:     paramPort.Port,
			Protocol: paramPort.Protocol,
		}

		if paramPort.HttpProxy != nil {
			port.HttpProxy = &model.PortHttpProxy{
				Name: paramPort.HttpProxy.Name,
			}

			if paramPort.HttpProxy.CustomDomain != nil {
				port.HttpProxy.CustomDomain = &model.CustomDomain{
					Domain: *paramPort.HttpProxy.CustomDomain,
					Secret: generateCustomDomainSecret(),
					Status: model.CustomDomainStatusPending,
				}
			}
		}

		portMap[paramPort.Name] = port
	}

	vm := model.VM{
		ID:      id,
		Name:    params.Name,
		Version: version.V2,
		Zone:    params.Zone,
		OwnerID: owner,

		CreatedAt:  time.Now(),
		UpdatedAt:  time.Time{},
		RepairedAt: time.Time{},
		DeletedAt:  time.Time{},
		AccessedAt: time.Now(),

		SshPublicKey: params.SshPublicKey,
		PortMap:      portMap,
		Specs: model.VmSpecs{
			CpuCores: params.CpuCores,
			RAM:      params.RAM,
			DiskSize: params.DiskSize,
		},

		Subsystems: model.Subsystems{},
		Activities: map[string]model.Activity{model.ActivityBeingCreated: {Name: model.ActivityBeingCreated, CreatedAt: time.Now()}},

		Host:   nil,
		Status: status_codes.GetMsg(status_codes.ResourceCreating),
	}

	_, err := client.Collection.InsertOne(context.TODO(), vm)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, rErrors.ErrNonUniqueField
		}

		return nil, fmt.Errorf("failed to create vm %s. details: %w", id, err)
	}

	return client.GetByID(id)
}

// ListWithAnyPendingCustomDomain returns a list of VMs that have any port with a pending custom domain.
// It uses aggregation to do this, so it is not very efficient.
func (client *Client) ListWithAnyPendingCustomDomain() ([]model.VM, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$addFields",
			Value: bson.D{{
				Key: "portMapArray",
				Value: bson.D{{
					Key:   "$objectToArray",
					Value: "$portMap",
				}},
			}},
		}},
		{{Key: "$match",
			Value: bson.D{{
				Key: "portMapArray",
				Value: bson.D{{
					Key: "$elemMatch",
					Value: bson.D{{
						Key: "v.httpProxy.customDomain.status",
						Value: bson.D{{
							Key:   "$in",
							Value: []string{model.CustomDomainStatusPending, model.CustomDomainStatusVerificationFailed},
						}},
					}},
				}},
			}},
		}},
	}

	cursor, err := client.Collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []model.VM{}, nil
		}

		return nil, err
	}

	var vms []model.VM
	err = cursor.All(context.Background(), &vms)
	if err != nil {
		return nil, err
	}

	var filteredVms []model.VM
	for _, vm := range vms {
		if vm.DoingActivity(model.ActivityBeingDeleted) {
			continue
		}

		filteredVms = append(filteredVms, vm)
	}

	return filteredVms, nil
}

func (client *Client) UpdateWithParams(id string, params *model.VmUpdateParams) error {
	setUpdate := bson.D{}
	unsetUpdate := bson.D{}

	// Check that no other vm has a port with the same http proxy name.
	// This should ideally be done with unique indexes, but the indexes are per document, and not per-array element.
	// So we would need to create a separate collection for ports, which is not ideal, since that would require
	// us to use $lookup to get the ports for a vm or use transactions, which is quite annoying :)
	//
	// So... we do it manually here, even though it might cause a race condition, because I'm lazy.
	if params.PortMap != nil {
		for _, port := range *params.PortMap {
			if port.HttpProxy != nil {
				filter := bson.D{
					{Key: "id", Value: bson.D{{Key: "$ne", Value: id}}},
					{Key: "ports.httpProxy.name", Value: port.HttpProxy.Name},
				}

				existAny, err := client.ResourceClient.ExistsWithFilter(filter)
				if err != nil {
					return err
				}

				if existAny {
					return rErrors.ErrNonUniqueField
				}
			}
		}
	}

	// Updating ports requires some extra love!
	// (since we delete custom domains by setting them to "")
	if params.PortMap != nil {
		onlyPorts, err := client.GetWithFilterAndProjection(bson.D{{Key: "id", Value: id}}, bson.D{{Key: "portMap", Value: 1}})
		if err != nil {
			return err
		}

		if onlyPorts.PortMap == nil {
			onlyPorts.PortMap = make(map[string]model.Port)
		}

		for mapName, port := range *params.PortMap {
			// Remove from onlyPorts
			delete(onlyPorts.PortMap, mapName)

			db.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.name", mapName), port.Name)
			db.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.port", mapName), port.Port)
			db.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.protocol", mapName), port.Protocol)

			if port.HttpProxy != nil {
				db.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.httpProxy.name", mapName), port.HttpProxy.Name)

				if port.HttpProxy.CustomDomain != nil {
					if *port.HttpProxy.CustomDomain == "" {
						db.Add(&unsetUpdate, fmt.Sprintf("portMap.%s.httpProxy.customDomain", mapName), "")
					} else {
						db.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.httpProxy.customDomain", mapName), model.CustomDomain{
							Domain: *port.HttpProxy.CustomDomain,
							Secret: generateCustomDomainSecret(),
							Status: model.CustomDomainStatusPending,
						})
					}
				}
			} else {
				db.Add(&unsetUpdate, fmt.Sprintf("portMap.%s.httpProxy", mapName), "")
			}
		}

		// Remove ports that are not in params.PortMap
		for name := range onlyPorts.PortMap {
			db.Add(&unsetUpdate, fmt.Sprintf("portMap.%s", name), "")
		}
	}

	db.AddIfNotNil(&setUpdate, "name", params.Name)
	db.AddIfNotNil(&setUpdate, "ownerId", params.OwnerID)
	db.AddIfNotNil(&setUpdate, "specs.cpuCores", params.CpuCores)
	db.AddIfNotNil(&setUpdate, "specs.ram", params.RAM)

	err := client.UpdateWithBsonByID(id,
		bson.D{
			{Key: "$set", Value: setUpdate},
			{Key: "$unset", Value: unsetUpdate},
		},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return rErrors.ErrNonUniqueField
		}

		return fmt.Errorf("failed to update vm %s. details: %w", id, err)
	}

	return nil
}

// GetUsage returns the usage in CPU cores, RAM, disk size and snapshots.
func (client *Client) GetUsage() (*model.VmUsage, error) {
	projection := bson.D{
		{Key: "_id", Value: 0},
		{Key: "id", Value: 1},
		{Key: "name", Value: 1},
		{Key: "specs", Value: 1},
	}

	vms, err := client.ListWithFilterAndProjection(bson.D{}, projection)
	if err != nil {
		return nil, err
	}

	usage := &model.VmUsage{
		CpuCores: 0,
		RAM:      0,
		DiskSize: 0,
	}

	for _, vm := range vms {
		usage.CpuCores += vm.Specs.CpuCores
		usage.RAM += vm.Specs.RAM
		usage.DiskSize += vm.Specs.DiskSize
	}

	return usage, nil
}

// DeleteSubsystem erases a subsystem from a VM.
// It prepends the key with `subsystems` and unsets it.
func (client *Client) DeleteSubsystem(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{Key: "$unset", Value: bson.D{{Key: subsystemKey, Value: ""}}}})
}

// SetSubsystem sets a subsystem in a VM.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{Key: subsystemKey, Value: update}})
}

// UpdateCustomDomainStatus updates the status of a custom domain for a given port.
func (client *Client) UpdateCustomDomainStatus(id, portName, status string) error {
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "portMap." + portName + ".httpProxy.customDomain.status", Value: status}}},
	}

	return client.UpdateWithBsonByID(id, update)
}

// SetStatusByName sets the status of a deployment.
func (client *Client) SetStatusByName(name, status string) error {
	return client.SetWithBsonByName(name, bson.D{{Key: "status", Value: status}})
}

// SetCurrentHost sets the current host of a VM.
func (client *Client) SetCurrentHost(name string, host *model.VmHost) error {
	return client.SetWithBsonByName(name, bson.D{{Key: "host", Value: host}})
}

// UnsetCurrentHost unsets the current host of a VM.
func (client *Client) UnsetCurrentHost(name string) error {
	return client.UnsetByName(name, "host")
}

// MarkRepaired marks a VM as repaired.
// It sets RepairedAt and unsets the repairing activity.
func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{Key: "id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "repairedAt", Value: time.Now()}}},
		{Key: "$unset", Value: bson.D{{Key: "activities.repairing", Value: ""}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// MarkUpdated marks a VM as updated.
// It sets UpdatedAt and unsets the updating activity.
func (client *Client) MarkUpdated(id string) error {
	filter := bson.D{{Key: "id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "updatedAt", Value: time.Now()}}},
		{Key: "$unset", Value: bson.D{{Key: "activities.updating", Value: ""}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// MarkAccessed marks a deployment as accessed to the current time.
func (client *Client) MarkAccessed(id string) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "accessedAt", Value: time.Now()}})
}

// generateCustomDomainSecret generates a random alphanumeric string.
func generateCustomDomainSecret() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
