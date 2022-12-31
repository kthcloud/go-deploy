package pdns

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/pdns/models"
	"log"
)

func ExampleCreate() {
	client, err := New(&ClientConf{
		ApiUrl: conf.Env.PDNS.Url,
		ApiKey: conf.Env.PDNS.Key,
		Zone:   conf.Env.PDNS.Zone,
	})

	if err != nil {
		log.Fatalln(err)
	}

	id, err := client.CreateRecord(&models.RecordPublic{
		Hostname:   "test",
		RecordType: "CNAME",
		TTL:        60,
		Content:    []string{client.Zone},
	})

	if err != nil {
		log.Fatalln(err)
	}

	if id == "" {
		log.Fatalln("no id received")
	}

	record, err := client.ReadRecord(id)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("received", record)

	err = client.DeleteRecord(record.ID)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("deleted", record.ID)
}
