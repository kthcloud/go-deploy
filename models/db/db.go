package db

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

var DB Context

type Context struct {
	mongoClient *mongo.Client
	RedisClient *redis.Client

	CollectionMap map[string]*mongo.Collection
}

func (dbCtx *Context) GetCollection(collectionName string) *mongo.Collection {
	if dbCtx.CollectionMap == nil || dbCtx.CollectionMap[collectionName] == nil {
		log.Fatalln("collection " + collectionName + " not found")
	}

	return dbCtx.CollectionMap[collectionName]
}

func Setup() {
	DB = Context{
		CollectionMap: make(map[string]*mongo.Collection),
	}

	err := DB.setupMongo()
	if err != nil {
		log.Fatalln(err)
	}

	err = DB.setupRedis()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("successfully setup db")
}

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
