package v2

import (
	"fmt"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func extractJobID(job *model.Job) (string, error) {
	if job == nil {
		return "", fmt.Errorf("missing job")
	}
	if err := utils.AssertParameters(job, []string{"id"}); err != nil {
		return "", err
	}
	id, ok := job.Args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("invalid id parameter")
	}
	return id, nil
}

func extractIDAndParams(job *model.Job) (string, map[string]any, error) {
	id, err := extractJobID(job)
	if err != nil {
		return "", nil, err
	}
	if err := utils.AssertParameters(job, []string{"params"}); err != nil {
		return "", nil, err
	}

	switch p := job.Args["params"].(type) {
	case map[string]any:
		if p == nil {
			return "", nil, fmt.Errorf("params is nil")
		}
		return id, p, nil
	case bson.M:
		if p == nil {
			return "", nil, fmt.Errorf("params is nil")
		}
		return id, map[string]any(p), nil
	default:
		return "", nil, fmt.Errorf("invalid params type %T", job.Args["params"])
	}
}
