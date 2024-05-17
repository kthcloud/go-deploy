package vm_repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/db"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	// NonUniqueFieldErr is returned when a unique index in the database is violated.
	NonUniqueFieldErr = fmt.Errorf("non unique field")
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
		Version: params.Version,

		OwnerID: owner,

		Zone:           params.Zone,
		DeploymentZone: params.DeploymentZone,

		CreatedAt:  time.Now(),
		UpdatedAt:  time.Time{},
		RepairedAt: time.Time{},
		DeletedAt:  time.Time{},
		AccessedAt: time.Now(),

		SshPublicKey: params.SshPublicKey,
		PortMap:      portMap,
		Activities:   map[string]model.Activity{model.ActivityBeingCreated: {model.ActivityBeingCreated, time.Now()}},
		Subsystems:   model.Subsystems{},
		Specs: model.VmSpecs{
			CpuCores: params.CpuCores,
			RAM:      params.RAM,
			DiskSize: params.DiskSize,
		},

		Status: status_codes.GetMsg(status_codes.ResourceCreating),
	}

	_, err := client.Collection.InsertOne(context.TODO(), vm)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, NonUniqueFieldErr
		}

		return nil, fmt.Errorf("failed to create vm %s. details: %w", id, err)
	}

	return client.GetByID(id)
}

// ListWithAnyPendingCustomDomain returns a list of VMs that have any port with a pending custom domain.
// It uses aggregation to do this, so it is not very efficient.
func (client *Client) ListWithAnyPendingCustomDomain() ([]model.VM, error) {
	pipeline := mongo.Pipeline{
		{{"$addFields", bson.D{{"portMapArray", bson.D{{"$objectToArray", "$portMap"}}}}}},
		{{"$match", bson.D{{"portMapArray", bson.D{{"$elemMatch", bson.D{{"v.httpProxy.customDomain.status", bson.D{{"$in", []string{model.CustomDomainStatusPending, model.CustomDomainStatusVerificationFailed}}}}}}}}}}},
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
					{"id", bson.D{{"$ne", id}}},
					{"ports.httpProxy.name", port.HttpProxy.Name},
				}

				existAny, err := client.ResourceClient.ExistsWithFilter(filter)
				if err != nil {
					return err
				}

				if existAny {
					return NonUniqueFieldErr
				}
			}
		}
	}

	// Updating ports requires some extra love!
	// (since we delete custom domains by setting them to "")
	if params.PortMap != nil {
		onlyPorts, err := client.GetWithFilterAndProjection(bson.D{{"id", id}}, bson.D{{"portMap", 1}})
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
			{"$set", setUpdate},
			{"$unset", unsetUpdate},
		},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return NonUniqueFieldErr
		}

		return fmt.Errorf("failed to update vm %s. details: %w", id, err)
	}

	return nil
}

// GetUsage returns the usage in CPU cores, RAM, disk size and snapshots.
func (client *Client) GetUsage() (*model.VmUsage, error) {
	projection := bson.D{
		{"_id", 0},
		{"id", 1},
		{"name", 1},
		{"specs", 1},
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
	return client.UpdateWithBsonByID(id, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

// SetSubsystem sets a subsystem in a VM.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}

// UpdateCustomDomainStatus updates the status of a custom domain for a given port.
func (client *Client) UpdateCustomDomainStatus(id, portName, status string) error {
	update := bson.D{
		{"$set", bson.D{{"portMap." + portName + ".httpProxy.customDomain.status", status}}},
	}

	return client.UpdateWithBsonByID(id, update)
}

// MarkRepaired marks a VM as repaired.
// It sets RepairedAt and unsets the repairing activity.
func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"repairedAt", time.Now()}}},
		{"$unset", bson.D{{"activities.repairing", ""}}},
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
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"updatedAt", time.Now()}}},
		{"$unset", bson.D{{"activities.updating", ""}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// MarkAccessed marks a deployment as accessed to the current time.
func (client *Client) MarkAccessed(id string) error {
	return client.SetWithBsonByID(id, bson.D{{"accessedAt", time.Now()}})
}

// generateCustomDomainSecret generates a random alphanumeric string.
func generateCustomDomainSecret() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
