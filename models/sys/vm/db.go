package vm

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models"
	"go-deploy/models/sys/activity"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	NonUniqueFieldErr = fmt.Errorf("non unique field")
)

func (client *Client) Create(id, owner, manager string, params *CreateParams) (*VM, error) {
	var portMap map[string]Port
	if params.PortMap != nil {
		portMap = make(map[string]Port)
		for _, paramPort := range params.PortMap {
			port := Port{
				Name:     paramPort.Name,
				Port:     paramPort.Port,
				Protocol: paramPort.Protocol,
			}

			if paramPort.HttpProxy != nil {
				port.HttpProxy = &PortHttpProxy{
					Name: paramPort.HttpProxy.Name,
				}

				if paramPort.HttpProxy.CustomDomain != nil {
					port.HttpProxy.CustomDomain = &CustomDomain{
						Domain: *paramPort.HttpProxy.CustomDomain,
						Secret: generateCustomDomainSecret(),
						Status: CustomDomainStatusPending,
					}
				}
			}

			portMap[paramPort.Name] = port
		}
	} else {
		portMap = make(map[string]Port)
	}

	vm := VM{
		ID:        id,
		Name:      params.Name,
		OwnerID:   owner,
		ManagedBy: manager,

		Zone:           params.Zone,
		DeploymentZone: params.DeploymentZone,

		CreatedAt:  time.Now(),
		UpdatedAt:  time.Time{},
		RepairedAt: time.Time{},
		DeletedAt:  time.Time{},

		SshPublicKey: params.SshPublicKey,
		PortMap:      map[string]Port{},
		Activities:   map[string]activity.Activity{ActivityBeingCreated: {ActivityBeingCreated, time.Now()}},
		Subsystems:   Subsystems{},
		Specs: Specs{
			CpuCores: params.CpuCores,
			RAM:      params.RAM,
			DiskSize: params.DiskSize,
		},

		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
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

func (client *Client) GetByTransferCode(code string) (*VM, error) {
	return client.GetWithFilterAndProjection(bson.D{{"transfer.code", code}}, nil)
}

func (client *Client) ListWithAnyPendingCustomDomain() ([]VM, error) {
	// Use aggregation to filter out vms that don't have any pending custom domains
	//db.collection.aggregate([
	//  {
	//	$addFields: {
	//	  portMapArray: { $objectToArray: "$portMap" }
	//	}
	//  },
	//  {
	//	$match: {
	//	  portMapArray: {
	//		$elemMatch: {
	//		  "v.httpProxy.customDomain.status": {
	//			$in: ["pending", "failed"]
	//		  }
	//		}
	//	  }
	//	}
	//  }
	//]);

	pipeline := mongo.Pipeline{
		{{"$addFields", bson.D{{"portMapArray", bson.D{{"$objectToArray", "$portMap"}}}}}},
		{{"$match", bson.D{{"portMapArray", bson.D{{"$elemMatch", bson.D{{"v.httpProxy.customDomain.status", bson.D{{"$in", []string{CustomDomainStatusPending, CustomDomainStatusVerificationFailed}}}}}}}}}}},
	}

	cursor, err := client.Collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []VM{}, nil
		}

		return nil, err
	}

	var vms []VM
	err = cursor.All(context.Background(), &vms)
	if err != nil {
		return nil, err
	}

	return vms, nil
}

func (client *Client) DeleteByID(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{
			{"$set", bson.D{{"deletedAt", time.Now()}}},
			{"$set", bson.D{{"activities", make(map[string]activity.Activity)}}},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s. details: %w", id, err)
	}

	return nil
}

func (client *Client) UpdateWithParamsByID(id string, params *UpdateParams) error {
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
			onlyPorts.PortMap = make(map[string]Port)
		}

		for mapName, port := range *params.PortMap {
			// Remove from onlyPorts
			delete(onlyPorts.PortMap, mapName)

			models.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.name", mapName), port.Name)
			models.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.port", mapName), port.Port)
			models.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.protocol", mapName), port.Protocol)

			if port.HttpProxy != nil {
				models.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.httpProxy.name", mapName), port.HttpProxy.Name)

				if port.HttpProxy.CustomDomain != nil {
					if *port.HttpProxy.CustomDomain == "" {
						models.Add(&unsetUpdate, fmt.Sprintf("portMap.%s.httpProxy.customDomain", mapName), "")
					} else {
						models.AddIfNotNil(&setUpdate, fmt.Sprintf("portMap.%s.httpProxy.customDomain", mapName), CustomDomain{
							Domain: *port.HttpProxy.CustomDomain,
							Secret: generateCustomDomainSecret(),
							Status: CustomDomainStatusPending,
						})
					}
				}
			} else {
				models.Add(&unsetUpdate, fmt.Sprintf("portMap.%s.httpProxy", mapName), "")
			}
		}

		// Remove ports that are not in params.PortMap
		for name := range onlyPorts.PortMap {
			models.Add(&unsetUpdate, fmt.Sprintf("portMap.%s", name), "")
		}
	}

	// If the transfer code is empty, it means it is either done and we remove it,
	// or we don't want to transfer it anymore
	if utils.EmptyValue(params.TransferCode) && utils.EmptyValue(params.TransferUserID) {
		models.AddIfNotNil(&unsetUpdate, "transfer", "")
	} else {
		models.AddIfNotNil(&setUpdate, "transfer.code", params.TransferCode)
		models.AddIfNotNil(&setUpdate, "transfer.userId", params.TransferUserID)
	}

	models.AddIfNotNil(&setUpdate, "name", params.Name)
	models.AddIfNotNil(&setUpdate, "ownerId", params.OwnerID)
	models.AddIfNotNil(&setUpdate, "specs.cpuCores", params.CpuCores)
	models.AddIfNotNil(&setUpdate, "specs.ram", params.RAM)

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

func (client *Client) DeleteSubsystemByID(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

func (client *Client) UpdateSubsystemByID(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}

func (client *Client) UpdateCustomDomainStatus(id, portName, status string) error {
	update := bson.D{
		{"$set", bson.D{{"portMap." + portName + ".httpProxy.customDomain.status", status}}},
	}

	return client.UpdateWithBsonByID(id, update)
}

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

func generateCustomDomainSecret() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
