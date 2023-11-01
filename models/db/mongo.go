package db

import (
	"context"
	"fmt"
	"go-deploy/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type CollectionDefinition struct {
	Name            string
	Indexes         []string
	UniqueIndexes   []string
	TextIndexFields []string
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

	DB.CollectionDefinitionMap = getCollectionDefinitions()

	for _, def := range DB.CollectionDefinitionMap {
		DB.CollectionMap[def.Name] = dbCtx.mongoClient.Database(config.Config.MongoDB.Name).Collection(def.Name)
	}

	log.Println("successfully found", len(DB.CollectionDefinitionMap), "collections")

	createdCount := 0
	for _, def := range DB.CollectionDefinitionMap {
		for _, indexName := range def.Indexes {
			_, err = DB.GetCollection(def.Name).Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys:    map[string]int{indexName: 1},
				Options: options.Index().SetUnique(false),
			})
			if err != nil {
				return makeError(err)
			}
			createdCount++
		}
	}

	log.Println("ensured", createdCount, "indexes")

	createdCount = 0
	for _, def := range DB.CollectionDefinitionMap {
		for _, indexName := range def.UniqueIndexes {
			_, err = DB.GetCollection(def.Name).Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys:    map[string]int{indexName: 1},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				return makeError(err)
			}
			createdCount++
		}
	}

	log.Println("ensured", createdCount, "unique indexes")

	createdCount = 0
	for _, def := range DB.CollectionDefinitionMap {
		if def.TextIndexFields == nil {
			continue
		}

		keys := bson.D{}
		for _, indexName := range def.TextIndexFields {
			keys = append(keys, bson.E{Key: indexName, Value: "text"})
		}

		_, err = DB.GetCollection(def.Name).Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: keys,
		})
		if err != nil {
			return makeError(err)
		}
		createdCount++
	}

	log.Println("ensured", createdCount, "text indexes")

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

func getCollectionDefinitions() map[string]CollectionDefinition {
	return map[string]CollectionDefinition{
		"deployments": {
			Name:          "deployments",
			Indexes:       []string{"name", "ownerId", "type", "statusCode", "createdAt", "deletedAt", "repairedAt", "restartedAt", "zone"},
			UniqueIndexes: []string{"id"},
		},
		"storageManagers": {
			Name:          "storageManagers",
			Indexes:       []string{"name", "ownerId", "createdAt", "deletedAt", "repairedAt", "zone"},
			UniqueIndexes: []string{"id"},
		},
		"vms": {
			Name:          "vms",
			Indexes:       []string{"name", "ownerId", "gpuId", "statusCode", "createdAt", "deletedAt", "repairedAt", "restartedAt", "zone"},
			UniqueIndexes: []string{"id"},
		},
		"gpus": {
			Name:          "gpus",
			Indexes:       []string{"name", "host", "lease.vmId", "lease.user", "lease.end"},
			UniqueIndexes: []string{"id"},
		},
		"users": {
			Name:            "users",
			Indexes:         []string{"username", "email", "firstName", "lastName", "effectiveRole.name"},
			UniqueIndexes:   []string{"id"},
			TextIndexFields: []string{"username", "email", "firstName", "lastName"},
		},
		"teams": {
			Name:            "teams",
			Indexes:         []string{"name", "ownerId", "createdAt", "deletedAt"},
			UniqueIndexes:   []string{"id"},
			TextIndexFields: []string{"name"},
		},
		"jobs": {
			Name:          "jobs",
			Indexes:       []string{"userId", "type", "args.id", "status", "createdAt", "runAfter"},
			UniqueIndexes: []string{"id"},
		},
		"notifications": {
			Name:          "notifications",
			Indexes:       []string{"userId", "type", "createdAt", "readAt", "deletedAt"},
			UniqueIndexes: []string{"id"},
		},
		"events": {
			Name:          "events",
			Indexes:       []string{"type", "createdAt", "source.userId"},
			UniqueIndexes: []string{"id"},
		},
	}
}
