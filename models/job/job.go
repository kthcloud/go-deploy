package job

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	JobCreateVM         = "createVm"
	JobDeleteVM         = "deleteVm"
	JobCreateDeployment = "createDeployment"
	JobDeleteDeployment = "deleteDeployment"
)

type Job struct {
	ID        string                 `bson:"id" json:"id"`
	UserID    string                 `bson:"userId" json:"userId"`
	Type      string                 `bson:"type" json:"type"`
	Args      map[string]interface{} `bson:"args" json:"args"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
	Status    string                 `bson:"status" json:"status"`
	ErrorLogs []string               `bson:"errorLogs" json:"errorLogs"`
}

func CreateJob(id, userID, jobType string, args map[string]interface{}) error {
	currentJob, err := GetJobByID(id)
	if err != nil {
		return err
	}

	if currentJob != nil {
		return fmt.Errorf("job with id %s already exists", id)
	}

	job := Job{
		ID:        id,
		UserID:    userID,
		Type:      jobType,
		Args:      args,
		CreatedAt: time.Now(),
		Status:    "pending",
	}

	_, err = models.JobCollection.InsertOne(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job. details: %s", err)
	}

	return nil
}

func GetJobByID(id string) (*Job, error) {
	var job Job
	err := models.JobCollection.FindOne(context.TODO(), bson.D{{"id", id}}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch vm. details: %s", err)
		return nil, err
	}

	return &job, err
}

func GetNextJob() (*Job, error) {
	filter := bson.D{{"status", "pending"}}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", "processing"}}}}

	var job Job
	err := models.JobCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func CompleteJob(jobID string) error {
	filter := bson.D{{"id", jobID}}
	update := bson.D{{"$set", bson.D{{"status", "completed"}}}}

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %s", err)
	}

	return nil
}

func FailJob(jobID string, errorLogs []string) error {
	filter := bson.D{{"id", jobID}}
	update := bson.D{{"$set", bson.D{{"status", "failed"}, {"errorLogs", errorLogs}}}}

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %s", err)
	}

	return nil
}

func ResetProcessingJobs() error {
	filter := bson.D{{"status", "processing"}}
	update := bson.D{{"$set", bson.D{{"status", "pending"}}}}

	_, err := models.JobCollection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %s", err)
	}

	return nil
}
