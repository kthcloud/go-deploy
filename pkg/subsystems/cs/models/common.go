package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"sort"
	"time"
)

type Tag struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

// FromCsTags converts the CloudStack tags to the internal tags.
func FromCsTags(tags []cloudstack.Tags) []Tag {
	var result []Tag
	for _, tag := range tags {
		result = append(result, Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result
}

// formatCreatedAt converts the CloudStack created time string to a time.Time.
func formatCreatedAt(created string) time.Time {
	iso8601 := "2006-01-02T15:04:05Z0700"
	createdAt, err := time.Parse(iso8601, created)
	if err != nil {
		return time.Now()
	}

	return createdAt
}
