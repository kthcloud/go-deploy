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
var VmCollection *mongo.Collection
var GpuCollection *mongo.Collection
var UserCollection *mongo.Collection
var JobCollection *mongo.Collection

var client *mongo.Client

func getUri() string {
	// this function is kept to allow easy switch from connString -> username + password + url etc.
	return conf.Env.DB.Url
}

func Setup() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup database. details: %s", err)
	}

	// Connect to db
	uri := getUri()
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

	log.Println("successfully connected to database")

	// Find collections
	DeploymentCollection = findCollection("deployments")
	VmCollection = findCollection("vms")
	GpuCollection = findCollection("gpus")
	UserCollection = findCollection("users")
	JobCollection = findCollection("jobs")
}

func findCollection(collectionName string) *mongo.Collection {
	collection := client.Database(conf.Env.DB.Name).Collection(collectionName)
	log.Println("found collection " + collectionName)
	return collection
}

func Shutdown() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to shutdown database. details: %s", err)
	}

	DeploymentCollection = nil
	VmCollection = nil
	GpuCollection = nil
	UserCollection = nil
	JobCollection = nil

	err := client.Disconnect(context.Background())
	if err != nil {
		log.Fatalln(makeError(err))
	}
}
