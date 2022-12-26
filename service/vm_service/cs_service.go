package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"log"
)

func CreateCS(name string) error {
	log.Println("setting up cs for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup k8s for v1_deployment %s. details: %s", name, err)
	}

	client, err := cs.New(&cs.ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		SecretKey: conf.Env.CS.Secret,
	})

	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByName(name)

	if err != nil {
		return makeError(err)
	}

	if len(vm.Subsystems.CS.VM.ID) == 0 {
		id, err := client.CreateVM(&csModels.VmPublic{
			Name: name,
			// temporary until vm templates are set up
			ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
			TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
			NetworkID:         "4a065a52-f290-4d2e-aeb4-6f48d3bd9bfe", // deploy
			ZoneID:            "3a74db73-6058-4520-8d8c-ab7d9b7955c8", // Flemingsberg
			ProjectID:         "d1ba382b-e310-445b-a54b-c4e773662af3", // deploy
		})
		if err != nil {
			return makeError(err)
		}

		csVM, err := client.ReadVM(id)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "cs", "vm", *csVM)
		if err != nil {
			return makeError(err)
		}
	}

	if err != nil {
		return makeError(err)
	}

	return nil
}

func DeleteCS(name string) error {
	log.Println("deleting npm for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup npm for v1_deployment %s. details: %s", name, err)
	}

	client, err := cs.New(&cs.ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		SecretKey: conf.Env.CS.Secret,
	})
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByName(name)

	if len(vm.Subsystems.CS.VM.ID) == 0 {
		return nil
	}

	err = client.DeleteVM(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.UpdateSubsystemByName(name, "cs", "vm", csModels.VmPublic{})
	if err != nil {
		return makeError(err)
	}

	return nil

}
