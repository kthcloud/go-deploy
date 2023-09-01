package job

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto/body"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (job *Job) ToDTO(statusMessage string) body.JobRead {
	var lastError *string
	if len(job.ErrorLogs) > 0 {
		lastError = &job.ErrorLogs[len(job.ErrorLogs)-1]
	}

	return body.JobRead{
		ID:        job.ID,
		UserID:    job.UserID,
		Type:      job.Type,
		Status:    statusMessage,
		LastError: lastError,
	}
}

func CreateJob(id, userID, jobType string, args map[string]interface{}) error {
	return CreateScheduledJob(id, userID, jobType, time.Now(), args)
}

func CreateScheduledJob(id, userID, jobType string, runAfter time.Time, args map[string]interface{}) error {
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
		RunAfter:  runAfter,
		Status:    StatusPending,
		ErrorLogs: make([]string, 0),
	}

	_, err = models.JobCollection.InsertOne(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job. details: %w", err)
	}

	return nil
}

func Exists(jobType string, args map[string]interface{}) (bool, error) {
	filter := bson.D{
		{"type", jobType},
		{"args", args},
		{"status", bson.D{{"$ne", StatusTerminated}}},
	}
	count, err := models.JobCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetByID(id string) (*Job, error) {
	var job Job
	err := models.JobCollection.FindOne(context.TODO(), bson.D{{"id", id}}).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch vm. details: %w", err)
		return nil, err
	}

	return &job, err
}

func GetNext() (*Job, error) {
	now := time.Now()
	filter := bson.D{
		{"status", StatusPending},
		{"runAfter", bson.D{{"$lte", now}}},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", StatusRunning}, {"lastRunAt", time.Now()}}}}

	var job Job
	err := models.JobCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &job, nil
}

func GetNextFailed() (*Job, error) {
	now := time.Now()
	filter := bson.D{
		{"status", StatusFailed},
		{"runAfter", bson.D{{"$lte", now}}},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{"createdAt", -1}})
	update := bson.D{{"$set", bson.D{{"status", StatusRunning}, {"lastRunAt", time.Now()}}}}

	var job Job
	err := models.JobCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &job, nil
}

func MarkCompleted(jobID string) error {
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

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

func MarkFailed(jobID string, reason string) error {
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

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

func MarkTerminated(jobID string, reason string) error {
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

	_, err := models.JobCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}

func ResetRunning() error {
	filter := bson.D{{"status", StatusRunning}}
	update := bson.D{{"$set", bson.D{{"status", StatusPending}}}}

	_, err := models.JobCollection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	err = CleanUp()
	if err != nil {
		return fmt.Errorf("failed to clean up job. details: %w", err)
	}

	return nil
}

func CleanUp() error {
	filter := bson.D{{"errorLogs", nil}}
	update := bson.D{{"$set", bson.D{{"errorLogs", make([]string, 0)}}}}

	_, err := models.JobCollection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job. details: %w", err)
	}

	return nil
}
