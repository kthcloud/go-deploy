package system

import (
	"context"
	"fmt"

	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	"github.com/kthcloud/go-deploy/pkg/db/resources/system_capacities_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/system_stats_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/system_status_repo"
)

// ListCapacities fetches the system capacities from the database.
func (c *Client) ListCapacities(ctx context.Context, n int) ([]body.TimestampedSystemCapacities, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch capacities. details: %s", err)
	}

	if n == 0 {
		n = 1
	}

	result, err := system_capacities_repo.New(n).WithContext(ctx).List()
	if err != nil {
		return nil, makeError(err)
	}

	return result, nil
}

// ListStats fetches the system stats from the database.
func (c *Client) ListStats(ctx context.Context, n int) ([]body.TimestampedSystemStats, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch stats. details: %s", err)
	}

	if n == 0 {
		n = 1
	}

	result, err := system_stats_repo.New(n).WithContext(ctx).List()
	if err != nil {
		return nil, makeError(err)
	}

	return result, nil
}

// ListStatus fetches the system status from the database.
func (c *Client) ListStatus(ctx context.Context, n int) ([]body.TimestampedSystemStatus, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch status. details: %s", err)
	}

	if n == 0 {
		n = 1
	}

	result, err := system_status_repo.New(n).WithContext(ctx).List()
	if err != nil {
		return nil, makeError(err)
	}

	return result, nil
}
