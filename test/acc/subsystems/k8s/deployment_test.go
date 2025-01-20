package k8s

import (
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/acc"
	"testing"
)

func TestCreateDeployment(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultDeployment(t, c)
}

func TestCreateNginxDeployment(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)

	d := &models.DeploymentPublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		Image:     "nginx:latest",
		EnvVars:   []models.EnvVar{{Name: acc.GenName(), Value: acc.GenName()}},
	}

	withDeployment(t, c, d)
}

func TestUpdateDeployment(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	d := withDefaultDeployment(t, c)

	d.EnvVars = []models.EnvVar{
		{Name: acc.GenName(), Value: acc.GenName()},
		{Name: acc.GenName(), Value: acc.GenName()},
	}
	d.Args = []string{acc.GenName()}

	dUpdated, err := c.UpdateDeployment(d)
	test.NoError(t, err, "failed to update deployment")

	test.EqualOrEmpty(t, d.EnvVars, dUpdated.EnvVars, "env vars do not match")
	test.EqualOrEmpty(t, d.Args, dUpdated.Args, "args do not match")
}
