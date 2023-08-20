package cs

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
)

func ExampleCreate() {
	zoneName := "Flemingsberg"
	zone := conf.Env.VM.GetZone(zoneName)

	client, err := New(&ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		ZoneID:      zone.ID,
		ProjectID:   zone.ProjectID,
		IpAddressID: zone.IpAddressID,
		NetworkID:   zone.NetworkID,
	})

	if err != nil {
		log.Fatalln(err)
	}

	id, err := client.CreateVM(&models.VmPublic{
		Name:              "demo",
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
		Tags:              []models.Tag{{Key: "managedBy", Value: "test"}},
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
	zoneName := "Flemingsberg"
	zone := conf.Env.VM.GetZone(zoneName)

	client, err := New(&ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		ZoneID:      zone.ID,
		ProjectID:   zone.ProjectID,
		IpAddressID: zone.IpAddressID,
		NetworkID:   zone.NetworkID,
	})

	if err != nil {
		log.Fatalln(err)
	}

	err = client.UpdateVM(&models.VmPublic{
		ID:                id,
		Name:              "demo",
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287", // Ubuntu Server
		ExtraConfig:       "extra config",
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("updated vm with id", id)
}
