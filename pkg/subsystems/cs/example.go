package cs

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
)

func Example() {
	client, _ := New(&ClientConf{
		ApiUrl:    conf.Env.CS.Url,
		ApiKey:    conf.Env.CS.Key,
		ApiSecret: conf.Env.CS.Secret,
		ZoneID:    conf.Env.CS.ZoneID,
	})

	params := &models.CreateVMParams{
		ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9",
		TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287",
		ZoneID:            conf.Env.CS.ZoneID,
	}

	err := client.CreateVM("deploy-hello", params)
	if err != nil {
		log.Fatalln(err)
	}
}
