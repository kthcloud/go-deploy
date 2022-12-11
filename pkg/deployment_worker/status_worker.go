package deployment_worker

import (
	"go-deploy/models"
	"go-deploy/pkg/app"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func Setup(ctx *app.Context) {
	log.Println("starting worker")
	go func() {
		for {
			if ctx.Stop {
				break
			}

			beingCreated, _ := models.GetAllDeploymentsWithFilter(bson.D{{"beingCreated", true}})
			for _, deployment := range beingCreated {
				created := Created(deployment.Name)
				if created {
					log.Printf("marking deployment %s as created\n", deployment.Name)
					_ = models.UpdateDeployment(deployment.ID, bson.D{{"beingCreated", false}})
				}
			}

			beingDeleted, _ := models.GetAllDeploymentsWithFilter(bson.D{{"beingDeleted", true}})
			for _, deployment := range beingDeleted {
				deleted := Deleted(deployment.Name)
				if deleted {
					log.Printf("deleting deployment %s\n", deployment.Name)
					_ = models.DeleteDeployment(deployment.ID, deployment.Owner)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
