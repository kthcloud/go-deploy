package harbor

import (
	"github.com/kthcloud/go-deploy/pkg/subsystems/harbor/models"
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/acc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateRobot(t *testing.T) {
	c, _ := withContext(t)
	withDefaultRobot(t, c)
}

func TestCreateRobotWithSecret(t *testing.T) {
	c, _ := withContext(t)
	r := withRobot(t, c, &models.RobotPublic{
		Name:   acc.GenName(),
		Secret: "Some-secret123",
	})

	assert.NotEmpty(t, r.Secret, "robot secret is empty")
}

func TestUpdateRobot(t *testing.T) {
	c, _ := withContext(t)
	r := withDefaultRobot(t, c)

	r.Disable = true

	rUpdated, err := c.UpdateRobot(r)
	test.NoError(t, err, "failed to update robot")

	assert.Equal(t, r.Disable, rUpdated.Disable, "robot disable is not updated")
}

func TestUpdateRobotWithSecret(t *testing.T) {
	c, _ := withContext(t)
	r := withDefaultRobot(t, c)

	r.Secret = "New-secret123"

	rUpdated, err := c.UpdateRobot(r)
	test.NoError(t, err, "failed to update robot")

	assert.Equal(t, r.Secret, rUpdated.Secret, "robot secret is not updated")
}

func TestUpdateRobotNewSecret(t *testing.T) {
	c, _ := withContext(t)
	r := withRobot(t, c, &models.RobotPublic{
		Name:   acc.GenName(),
		Secret: "Some-secret123",
	})

	r.Secret = "New-secret123"

	rUpdated, err := c.UpdateRobot(r)
	test.NoError(t, err, "failed to update robot")

	assert.Equal(t, r.Secret, rUpdated.Secret, "robot secret is not updated")
}
