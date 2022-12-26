package deployment_worker

import (
	deploymentModel "go-deploy/models/deployment"
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

			beingCreated, _ := deploymentModel.GetAllDeploymentsWithFilter(bson.D{{"beingCreated", true}})
			for _, deployment := range beingCreated {
				created := Created(&deployment)
				if created {
					log.Printf("marking deployment %s as created\n", deployment.Name)
					_ = deploymentModel.UpdateDeployment(deployment.ID, bson.D{{"beingCreated", false}})
				}
			}

			beingDeleted, _ := deploymentModel.GetAllDeploymentsWithFilter(bson.D{{"beingDeleted", true}})
			for _, deployment := range beingDeleted {
				deleted := Deleted(&deployment)
				if deleted {
					log.Printf("marking deployment %s as deleted\n", deployment.Name)
					_ = deploymentModel.DeleteDeployment(deployment.ID, deployment.Owner)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
