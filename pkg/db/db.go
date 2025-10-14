package db

import (
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
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

// Setup initializes the database context.
// It should be called once at the start of the application.
func Setup() error {
	DB = Context{
		CollectionMap: make(map[string]*mongo.Collection),
	}

	err := DB.setupMongo()
	if err != nil {
		return err
	}

	err = DB.setupRedis()
	if err != nil {
		return err
	}

	return nil
}

// Shutdown closes the database connections.
// It should be called once at the end of the application.
func Shutdown() {
	err := DB.shutdownRedis()
	if err != nil {
		log.Fatalln(err)
	}

	err = DB.shutdownMongo()
	if err != nil {
		log.Fatalln(err)
	}
}
