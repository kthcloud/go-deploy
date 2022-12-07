package worker

import (
	"go-deploy/models"
	"go-deploy/pkg/app"
	"go-deploy/pkg/subsystems"
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

			beingCreated, _ := models.GetAllProjectsWithCondition(bson.D{{"beingCreated", true}})
			for _, project := range beingCreated {
				created := subsystems.Created(project.Name)
				if created {
					log.Printf("marking project %s as created\n", project.Name)
					_ = models.UpdateProject(project.ID, bson.D{{"beingCreated", false}})
				}
			}

			beingDeleted, _ := models.GetAllProjectsWithCondition(bson.D{{"beingDeleted", true}})
			for _, project := range beingDeleted {
				deleted := subsystems.Deleted(project.Name)
				if deleted {
					log.Printf("deleting project %s\n", project.Name)
					_ = models.DeleteProject(project.ID, project.Owner)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
