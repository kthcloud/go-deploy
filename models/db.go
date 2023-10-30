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

var DB DbType

type DbType struct {
	CollectionMap map[string]*mongo.Collection
}

func (db *DbType) GetCollection(collectionName string) *mongo.Collection {
	if db.CollectionMap == nil || db.CollectionMap[collectionName] == nil {
		log.Fatalln("collection " + collectionName + " not found")
	}

	return db.CollectionMap[collectionName]
}

var client *mongo.Client

func getUri() string {
	// this function is kept to allow easy switch from connString -> username + password + url etc.
	return conf.Env.DB.URL
}

func Setup() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup database. details: %w", err)
	}

	// Connect to mongodb
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

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalln(makeError(err))
	}

	log.Println("successfully connected to database")

	// Find collections
	DB.CollectionMap = make(map[string]*mongo.Collection)

	collections := []string{
		"deployments",
		"storageManagers",
		"vms",
		"gpus",
		"users",
		"teams",
		"jobs",
		"notifications",
		"events",
	}

	for _, collectionName := range collections {
		log.Println("found collection " + collectionName)
		DB.CollectionMap[collectionName] = client.Database(conf.Env.DB.Name).Collection(collectionName)
	}

	// create unique indexes
	uniqueIndexes := map[string][]string{
		"deployments":     {"id"},
		"storageManagers": {"id", "ownerId"},
		"vms":             {"id"},
		"gpus":            {"id"},
		"users":           {"id"},
		"teams":           {"id"},
		"jobs":            {"id"},
		"notifications":   {"id"},
		"events":          {"id"},
	}

	for collectionName, fieldNames := range uniqueIndexes {
		for _, fieldName := range fieldNames {
			err = createUniqueIndex(DB.GetCollection(collectionName), fieldName)
			if err != nil {
				log.Fatalln(makeError(err))
			}
		}
	}
}

func createUniqueIndex(collection *mongo.Collection, fieldName string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create unique index on collection %s. details: %w", collection.Name(), err)
	}

	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    map[string]int{fieldName: 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Shutdown() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to shutdown database. details: %w", err)
	}

	DB.CollectionMap = nil

	err := client.Disconnect(context.Background())
	if err != nil {
		log.Fatalln(makeError(err))
	}
}
