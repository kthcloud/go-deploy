package job_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create creates a new job in the database.
func (client *Client) Create(id, userID, jobType, version string, args map[string]interface{}) error {
	return client.CreateScheduled(id, userID, jobType, version, time.Now(), args)
}

// CreateScheduled creates a new job in the database that will run after the given time.
func (client *Client) CreateScheduled(id, userID, jobType, version string, runAfter time.Time, args map[string]interface{}) error {
	currentJob, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if currentJob != nil {
		return fmt.Errorf("job with id %s already exists", id)
	}

	job := model.Job{
		ID:        id,
		UserID:    userID,
		Type:      jobType,
		Args:      args,
		Version:   version,
		CreatedAt: time.Now(),
		RunAfter:  runAfter,
		Attempts:  0,
		Status:    model.JobStatusPending,
		ErrorLogs: make([]string, 0),
	}

	_, err = client.Collection.InsertOne(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job. details: %w", err)
	}

	return nil
}

// GetNext returns the next job that should be executed.
func (client *Client) GetNext() (*model.Job, error) {
	now := time.Now()
	filter := bson.D{
		{Key: "status", Value: model.JobStatusPending},
		{Key: "runAfter", Value: bson.D{{Key: "$lte", Value: now}}},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: model.JobStatusRunning}, {Key: "lastRunAt", Value: time.Now()}}}}

	var job model.Job
	err := client.Collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &job, nil
}

// GetNextFailed returns the next job that failed and should be retried.
func (client *Client) GetNextFailed() (*model.Job, error) {
	now := time.Now()
	filter := bson.D{
		{Key: "status", Value: model.JobStatusFailed},
		{Key: "runAfter", Value: bson.D{{Key: "$lte", Value: now}}},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: model.JobStatusRunning}, {Key: "lastRunAt", Value: time.Now()}}}}

	var job model.Job
	err := client.Collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &job, nil
}

// MarkCompleted marks a job as completed.
func (client *Client) MarkCompleted(jobID string) error {
	filter := bson.D{{Key: "id", Value: jobID}}

	// update status and finishedAt
	update := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "status", Value: model.JobStatusCompleted},
				{Key: "finishedAt", Value: time.Now()},
			},
		},
	}

	_, err := client.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

// MarkFailed marks a job as failed, meaning it should be retried.
func (client *Client) MarkFailed(jobID string, runAfter time.Time, attempts int, reason string) error {
	filter := bson.D{
		{Key: "id", Value: jobID},
	}
	update := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "status", Value: model.JobStatusFailed},
				{Key: "finishedAt", Value: time.Now()},
				{Key: "attempts", Value: attempts},
				{Key: "runAfter", Value: runAfter},
			},
		},
		{Key: "$push",
			Value: bson.D{{Key: "errorLogs", Value: reason}},
		},
	}

	_, err := client.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

// MarkTerminated marks a job as terminated, meaning it should not be retried.
func (client *Client) MarkTerminated(jobID string, reason string) error {
	filter := bson.D{
		{Key: "id", Value: jobID},
	}
	update := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "status", Value: model.JobStatusTerminated},
				{Key: "finishedAt", Value: time.Now()},
			},
		},
		{Key: "$push",
			Value: bson.D{{Key: "errorLogs", Value: reason}},
		},
	}

	_, err := client.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

// ResetRunning resets all running jobs to pending.
// This is used when the application is restarted to prevent jobs from being stuck in running state.
func (client *Client) ResetRunning() error {
	filter := bson.D{{Key: "status", Value: model.JobStatusRunning}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: model.JobStatusPending}}}}

	_, err := client.Collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

// UpdateWithParams updates the job with the given params.
func (client *Client) UpdateWithParams(id string, params *model.JobUpdateParams) error {
	updateData := bson.D{}

	db.AddIfNotNil(&updateData, "status", params.Status)

	if len(updateData) == 0 {
		return nil
	}

	err := client.SetWithBsonByID(id, updateData)
	if err != nil {
		return fmt.Errorf("failed to update job %s. details: %w", id, err)
	}

	return nil
}
