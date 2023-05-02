package job

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto/body"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	TypeCreateVM         = "createVm"
	TypeDeleteVM         = "deleteVm"
	TypeCreateDeployment = "createDeployment"
	TypeDeleteDeployment = "deleteDeployment"
)

const (
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusFinished = "finished"
	StatusFailed   = "failed"
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

func (job *Job) ToDTO(statusMessage string) body.JobRead {
	if job == nil {
		return body.JobRead{}
	}

	return body.JobRead{
		ID:     job.ID,
		UserID: job.UserID,
		Type:   job.Type,
		Status: statusMessage,
	}
}

func CreateJob(id, userID, jobType string, args map[string]interface{}) error {
	currentJob, err := GetByID(id)
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
		Status:    StatusPending,
	}

	_, err = models.JobCollection.InsertOne(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job. details: %s", err)
	}

	return nil
}

func GetByID(id string) (*Job, error) {
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

func GetNext() (*Job, error) {
	filter := bson.D{{"status", StatusPending}}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", StatusRunning}}}}

	var job Job
	err := models.JobCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func GetNextFailed() (*Job, error) {
	filter := bson.D{{"status", StatusFailed}}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", StatusRunning}}}}

	var job Job
	err := models.JobCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func MarkCompleted(jobID string) error {
	filter := bson.D{{"id", jobID}}
	update := bson.D{{"$set", bson.D{{"status", StatusFinished}}}}

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %s", err)
	}

	return nil
}

func MarkFailed(jobID string, errorLogs []string) error {
	filter := bson.D{{"id", jobID}}
	update := bson.D{{"$set", bson.D{{"status", StatusFailed}, {"errorLogs", errorLogs}}}}

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %s", err)
	}

	return nil
}

func ResetRunning() error {
	filter := bson.D{{"status", StatusRunning}}
	update := bson.D{{"$set", bson.D{{"status", StatusPending}}}}

	_, err := models.JobCollection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %s", err)
	}

	return nil
}
