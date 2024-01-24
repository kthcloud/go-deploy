package job

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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

	job := Job{
		ID:        id,
		UserID:    userID,
		Type:      jobType,
		Args:      args,
		Version:   version,
		CreatedAt: time.Now(),
		RunAfter:  runAfter,
		Attempts:  0,
		Status:    StatusPending,
		ErrorLogs: make([]string, 0),
	}

	_, err = client.Collection.InsertOne(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job. details: %w", err)
	}

	return nil
}

// GetNext returns the next job that should be executed.
func (client *Client) GetNext() (*Job, error) {
	now := time.Now()
	filter := bson.D{
		{"status", StatusPending},
		{"runAfter", bson.D{{"$lte", now}}},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", StatusRunning}, {"lastRunAt", time.Now()}}}}

	var job Job
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
func (client *Client) GetNextFailed() (*Job, error) {
	now := time.Now()
	filter := bson.D{
		{"status", StatusFailed},
		{"runAfter", bson.D{{"$lte", now}}},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", StatusRunning}, {"lastRunAt", time.Now()}}}}

	var job Job
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
	filter := bson.D{{"id", jobID}}

	// update status and finishedAt
	update := bson.D{
		{"$set",
			bson.D{
				{"status", StatusCompleted},
				{"finishedAt", time.Now()},
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
		{"id", jobID},
	}
	update := bson.D{
		{"$set",
			bson.D{
				{"status", StatusFailed},
				{"finishedAt", time.Now()},
				{"attempts", attempts},
				{"runAfter", runAfter},
			},
		},
		{"$push",
			bson.D{{"errorLogs", reason}},
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
		{"id", jobID},
	}
	update := bson.D{
		{"$set",
			bson.D{
				{"status", StatusTerminated},
				{"finishedAt", time.Now()},
			},
		},
		{"$push",
			bson.D{{"errorLogs", reason}},
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
	filter := bson.D{{"status", StatusRunning}}
	update := bson.D{{"$set", bson.D{{"status", StatusPending}}}}

	_, err := client.Collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

// UpdateWithParams updates the job with the given params.
func (client *Client) UpdateWithParams(id string, params *UpdateParams) error {
	updateData := bson.D{}

	models.AddIfNotNil(&updateData, "status", params.Status)

	if len(updateData) == 0 {
		return nil
	}

	err := client.SetWithBsonByID(id, updateData)
	if err != nil {
		return fmt.Errorf("failed to update job %s. details: %w", id, err)
	}

	return nil
}
