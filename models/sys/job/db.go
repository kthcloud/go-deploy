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

func (client *Client) Create(id, userID, jobType string, args map[string]interface{}) error {
	return client.CreateScheduled(id, userID, jobType, time.Now(), args)
}

func (client *Client) CreateScheduled(id, userID, jobType string, runAfter time.Time, args map[string]interface{}) error {
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
		CreatedAt: time.Now(),
		RunAfter:  runAfter,
		Status:    StatusPending,
		ErrorLogs: make([]string, 0),
	}

	_, err = client.Collection.InsertOne(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job. details: %w", err)
	}

	return nil
}

func (client *Client) Exists(jobType string, args map[string]interface{}) (bool, error) {
	filter := bson.D{
		{"type", jobType},
		{"args", args},
		{"status", bson.D{{"$nin", []string{StatusCompleted, StatusTerminated}}}},
	}
	count, err := client.Collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *Client) GetMany(jobType, status *string) ([]Job, error) {
	filter := bson.D{}
	if jobType != nil {
		filter = append(filter, bson.E{Key: "type", Value: jobType})
	}

	if status != nil {
		filter = append(filter, bson.E{Key: "status", Value: status})
	}

	return client.ListWithFilter(filter)
}

func (client *Client) GetByArgs(args map[string]interface{}) ([]Job, error) {
	fullFilter := bson.D{}

	for key, value := range args {
		filter := bson.D{
			{"args." + key, value},
		}

		fullFilter = append(fullFilter, filter...)
	}

	return client.ListWithFilter(fullFilter)
}

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

func (client *Client) MarkFailed(jobID string, reason string) error {
	filter := bson.D{
		{"id", jobID},
	}
	update := bson.D{
		{"$set",
			bson.D{
				{"status", StatusFailed},
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

func (client *Client) ResetRunning() error {
	filter := bson.D{{"status", StatusRunning}}
	update := bson.D{{"$set", bson.D{{"status", StatusPending}}}}

	_, err := client.Collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	err = client.CleanUp()
	if err != nil {
		return fmt.Errorf("failed to clean up job. details: %w", err)
	}

	return nil
}

func (client *Client) CleanUp() error {
	filter := bson.D{{"errorLogs", nil}}
	update := bson.D{{"$set", bson.D{{"errorLogs", make([]string, 0)}}}}

	_, err := client.Collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

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
