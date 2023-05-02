package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

func AddIfNotNil(data bson.M, key string, value interface{}) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return
	}
	data[key] = value
}
