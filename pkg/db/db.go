package db

import (
	"context"
	"errors"

	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var DB Context

// Context is the database context for the application.
// It contains the mongo client and redis client, as well as
// a map of all collections and their definitions.
// It is used as a singleton, and should be initialized with
// the Setup() function.
type Context struct {
	MongoClient *mongo.Client
	RedisClient *redis.Client

	CollectionMap           map[string]*mongo.Collection
	CollectionDefinitionMap map[string]CollectionDefinition
}

// GetCollection returns the collection with the given name.
// If the collection is not found, the application will exit.
func (dbCtx *Context) GetCollection(collectionName string) *mongo.Collection {
	if dbCtx.CollectionMap == nil || dbCtx.CollectionMap[collectionName] == nil {
		log.Fatalln("collection " + collectionName + " not found")
	}

	return dbCtx.CollectionMap[collectionName]
}

// Setup initializes the database context using ctx for connection and index setup lifetimes.
// It should be called once at the start of the application.
func Setup(ctx context.Context, logger *zap.Logger) error {
	if ctx == nil {
		ctx = context.Background()
	}

	DB = Context{
		CollectionMap: make(map[string]*mongo.Collection),
	}

	if err := DB.setupMongo(ctx); err != nil {
		return err
	}

	if err := DB.setupRedis(ctx); err != nil {
		return err
	}

	return nil
}

// Shutdown closes the database connections using ctx for disconnect deadlines.
// It should be called once at the end of the application.
func Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	return errors.Join(
		DB.shutdownRedis(),
		DB.shutdownMongo(ctx),
	)
}
