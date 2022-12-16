package main

import (
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/npm"
	"go-deploy/pkg/terraform"
	"log"
	"path/filepath"
)

func Example() {

	workingDir, _ := filepath.Abs("terraform/deploy/demo")

	client, err := terraform.New(&terraform.ClientConf{
		WorkingDir: workingDir,
	})
	if err != nil {
		log.Fatalln(err)
	}

	client.Add(cs.GetTerraform())

	client.Add(npm.GetTerraform())

	err = client.Apply()
	if err != nil {
		log.Fatalln(err)
	}

	//client, err := New(&ClientConf{
	//	ApiUrl:       conf.Env.CS.Url,
	//	ApiKey:       conf.Env.CS.Key,
	//	SecretKey:    conf.Env.CS.Secret,
	//	ZoneID:       conf.Env.CS.ZoneID,
	//	TerraformDir: "C:\\repos\\go-deploy\\terraform",
	//})
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}
	//
	//_, err = client.Terraform.Show(context.TODO())
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}
	//
	//err = client.Terraform.Apply(context.TODO())
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}

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
