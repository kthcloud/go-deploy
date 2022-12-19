package models

import (
	"context"
	"fmt"
	"go-deploy/pkg/conf"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DeploymentCollection *mongo.Collection
var client *mongo.Client

func getUri() string {
	db := conf.Env.DB

	noCred := len(db.Username) == 0 || len(db.Password) == 0

	var url string
	if noCred {
		url = fmt.Sprintf("mongodb://%s", db.Url)
	} else {
		url = fmt.Sprintf("mongodb+srv://%s:%s@%s", db.Username, db.Password, db.Url)
	}

	return url
}

func Setup() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup models. details: %s", err)
	}

	// Connect to db
	uri := getUri()
	log.Println("connecting to database")
	clientResult, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalln(makeError(err))
	}
	client = clientResult

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalln(makeError(err))
	}

	// Find DeploymentCollection
	DeploymentCollection = client.Database(conf.Env.DB.Name).Collection("deployments")
	if err != nil {
		log.Fatalln(makeError(err))
	}
}

func Shutdown() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to shutdown models. details: %s", err)
	}

	err := client.Disconnect(context.Background())
	if err != nil {
		log.Fatalln(makeError(err))
	}
}
