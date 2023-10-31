package db

import (
	"context"
	"fmt"
	"go-deploy/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type collectionDefinition struct {
	Name          string
	Indexes       []string
	UniqueIndexes []string
}

func (dbCtx *Context) setupMongo() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup mongodb. details: %w", err)
	}

	log.Println("setting up mongodb")

	var err error
	dbCtx.mongoClient, err = mongo.NewClient(options.Client().ApplyURI(config.Config.MongoDB.URL))
	if err != nil {
		return makeError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = dbCtx.mongoClient.Connect(ctx)
	if err != nil {
		return makeError(err)
	}

	err = dbCtx.mongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalln(makeError(err))
	}

	log.Println("successfully connected to mongodb")

	// Find collections
	DB.CollectionMap = make(map[string]*mongo.Collection)

	defs := getCollectionDefinitions()

	for _, def := range defs {
		DB.CollectionMap[def.Name] = dbCtx.mongoClient.Database(config.Config.MongoDB.Name).Collection(def.Name)
	}

	log.Println("successfully found", len(defs), "collections")

	createdCount := 0
	for _, def := range defs {
		for _, indexName := range def.Indexes {
			err = createIndex(DB.GetCollection(def.Name), indexName, false)
			if err != nil {
				log.Fatalln(makeError(err))
			}
			createdCount++
		}
	}

	log.Println("ensured", createdCount, "indexes")

	createdCount = 0
	for _, def := range defs {
		for _, indexName := range def.UniqueIndexes {
			err = createIndex(DB.GetCollection(def.Name), indexName, true)
			if err != nil {
				log.Fatalln(makeError(err))
			}
			createdCount++
		}
	}

	log.Println("ensured", createdCount, "unique indexes")

	return nil
}

func (dbCtx *Context) shutdownMongo() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to shutdown database. details: %w", err)
	}

	err := dbCtx.mongoClient.Disconnect(context.Background())
	if err != nil {
		return makeError(err)
	}

	dbCtx.CollectionMap = nil

	return nil
}

func createIndex(collection *mongo.Collection, fieldName string, unique bool) error {
	uniqueStr := ""
	if unique {
		uniqueStr = "unique "
	}
	makeError := func(err error) error {
		return fmt.Errorf("failed to create %sindex on collection %s. details: %w", uniqueStr, collection.Name(), err)
	}

	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    map[string]int{fieldName: 1},
		Options: options.Index().SetUnique(unique),
	})
	if err != nil {
		return makeError(err)
	}

	return nil
}

func getCollectionDefinitions() []collectionDefinition {
	return []collectionDefinition{
		{
			Name:          "deployments",
			Indexes:       []string{"name", "ownerId", "type", "statusCode", "createdAt", "deletedAt", "repairedAt", "restartedAt", "zone"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "storageManagers",
			Indexes:       []string{"name", "ownerId", "createdAt", "deletedAt", "repairedAt", "zone"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "vms",
			Indexes:       []string{"name", "ownerId", "gpuId", "statusCode", "createdAt", "deletedAt", "repairedAt", "restartedAt", "zone"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "gpus",
			Indexes:       []string{"name", "host", "lease.vmId", "lease.user", "lease.end"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "users",
			Indexes:       []string{"username", "email", "firstName", "lastName", "effectiveRole.name"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "teams",
			Indexes:       []string{"name", "ownerId", "createdAt", "deletedAt"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "jobs",
			Indexes:       []string{"userId", "type", "args.id", "status", "createdAt", "runAfter"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "notifications",
			Indexes:       []string{"userId", "type", "createdAt", "readAt", "deletedAt"},
			UniqueIndexes: []string{"id"},
		},
		{
			Name:          "events",
			Indexes:       []string{"type", "createdAt", "source.userId"},
			UniqueIndexes: []string{"id"},
		},
	}
}
