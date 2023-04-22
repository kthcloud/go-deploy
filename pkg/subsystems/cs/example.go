package cs

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
)

func ExampleCreate() {
	client, err := New(&ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		SecretKey: conf.Env.CS.Secret,
	})

	if err != nil {
		log.Fatalln(err)
	}

	id, err := client.CreateVM(&models.VmPublic{
		Name:              "demo",
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
		NetworkID:         "4a065a52-f290-4d2e-aeb4-6f48d3bd9bfe", // deploy
		ZoneID:            "3a74db73-6058-4520-8d8c-ab7d9b7955c8", // Flemingsberg
		ProjectID:         "d1ba382b-e310-445b-a54b-c4e773662af3", // deploy
	},
		"public key 1", "public key 2",
	)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("created vm with id", id)
}

func ExampleUpdate() {
	id := "77b35b74-8333-4247-849d-ef5bc8555459"

	client, err := New(&ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		SecretKey: conf.Env.CS.Secret,
	})

	if err != nil {
		log.Fatalln(err)
	}

	err = client.UpdateVM(&models.VmPublic{
		ID:                id,
		Name:              "demo",
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
		NetworkID:         "4a065a52-f290-4d2e-aeb4-6f48d3bd9bfe", // deploy
		ZoneID:            "3a74db73-6058-4520-8d8c-ab7d9b7955c8", // Flemingsberg
		ProjectID:         "d1ba382b-e310-445b-a54b-c4e773662af3", // deploy
		ExtraConfig:       "extra config",
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("updated vm with id", id)
}
