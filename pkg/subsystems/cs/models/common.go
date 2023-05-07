package models

import "go-deploy/pkg/imp/cloudstack"

type Tag struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

func FromCsTags(tags []cloudstack.Tags) []Tag {
	var result []Tag
	for _, tag := range tags {
		result = append(result, Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	return result
}
