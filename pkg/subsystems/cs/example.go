package cs

import (
	"context"
	"go-deploy/pkg/conf"
	"log"
)

func Example() {
	client, err := New(&ClientConf{
		ApiUrl:       conf.Env.CS.Url,
		ApiKey:       conf.Env.CS.Key,
		SecretKey:    conf.Env.CS.Secret,
		ZoneID:       conf.Env.CS.ZoneID,
		TerraformDir: "C:\\repos\\go-deploy\\terraform",
	})
	if err != nil {
		log.Fatalln(err)
		return
	}

	_, err = client.Terraform.Show(context.TODO())
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = client.Terraform.Apply(context.TODO())
	if err != nil {
		log.Fatalln(err)
		return
	}

	//params := &models.CreateVMParams{
	//	ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9",
	//	TemplateID:        "e1a0479c-76a2-44da-8b38-a3a3fa316287",
	//	ZoneID:            conf.Env.CS.ZoneID,
	//}

	//err := client.CreateVM("deploy-hello", params)
	//if err != nil {
	//	log.Fatalln(err)
	//}
}
