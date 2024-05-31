package body

import "time"

type SystemStats struct {
	K8sStats K8sStats `json:"k8s" bson:"k8s"`
}

type TimestampedSystemStats struct {
	Stats     SystemStats `json:"stats" bson:"stats"`
	Timestamp time.Time   `json:"timestamp" bson:"timestamp"`
}

type K8sStats struct {
	PodCount int            `json:"podCount" bson:"podCount"`
	Clusters []ClusterStats `json:"clusters" bson:"clusters"`
}

type ClusterStats struct {
	Name     string `json:"cluster" bson:"cluster"`
	PodCount int    `json:"podCount" bson:"podCount"`
}
