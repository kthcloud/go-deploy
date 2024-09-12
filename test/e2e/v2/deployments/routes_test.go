package deployments

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/test/e2e"
	"github.com/kthcloud/go-deploy/test/e2e/v2"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestList(t *testing.T) {
	t.Parallel()

	queries := []string{
		"?page=1&pageSize=10",
		"?userId=" + model.TestPowerUserID + "&page=1&pageSize=3",
	}

	for _, query := range queries {
		v2.ListDeployments(t, query)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	requestBody := body.DeploymentCreate{
		Name: e2e.GenName(),
		Envs: []body.Env{
			{
				Name:  e2e.GenName(),
				Value: uuid.NewString(),
			},
		},
	}

	v2.WithDeployment(t, requestBody)
}

func TestCreateWithCustomPort(t *testing.T) {
	t.Parallel()

	// This test assumes that the default port is 8080
	customPort := 8081

	requestBody := body.DeploymentCreate{
		Name: e2e.GenName(),
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
	}

	v2.WithDeployment(t, requestBody)
}

func TestCreateWithCustomImage(t *testing.T) {
	t.Parallel()

	// This setup is chosen to make the deployment reachable (pingable)
	customImage := "nginx:latest"
	customPort := 80

	requestBody := body.DeploymentCreate{
		Name: e2e.GenName(),
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
		Image: &customImage,
	}

	v2.WithDeployment(t, requestBody)
}

func TestCreateWithCustomDomain(t *testing.T) {
	t.Parallel()

	customDomain := e2e.TestDomain

	requestBody := body.DeploymentCreate{
		Name:         e2e.GenName(),
		CustomDomain: &customDomain,
	}

	v2.WithDeployment(t, requestBody)
}

func TestCreateWithInvalidBody(t *testing.T) {
	t.Parallel()

	longName := body.DeploymentCreate{Name: "e2e-"}
	for i := 0; i < 1000; i++ {
		longName.Name += uuid.NewString()
	}
	v2.WithAssumedFailedDeployment(t, longName)

	invalidNames := []string{
		e2e.GenName() + "-",
		e2e.GenName() + "- ",
		e2e.GenName() + ".",
		"." + e2e.GenName(),
		e2e.GenName() + " " + e2e.GenName(),
		e2e.GenName() + "%",
		e2e.GenName() + "!",
		e2e.GenName() + "%" + e2e.GenName(),
	}

	for _, name := range invalidNames {
		requestBody := body.DeploymentCreate{Name: name}
		v2.WithAssumedFailedDeployment(t, requestBody)
	}

	tooManyEnvs := body.DeploymentCreate{Name: e2e.GenName()}
	tooManyEnvs.Envs = make([]body.Env, 10000)
	for i := range tooManyEnvs.Envs {
		tooManyEnvs.Envs[i] = body.Env{
			Name:  uuid.NewString(),
			Value: uuid.NewString(),
		}
	}

	v2.WithAssumedFailedDeployment(t, tooManyEnvs)

	tooManyVolumes := body.DeploymentCreate{Name: e2e.GenName()}
	tooManyVolumes.Volumes = make([]body.Volume, 10000)
	for i := range tooManyVolumes.Volumes {
		tooManyVolumes.Volumes[i] = body.Volume{
			Name:       uuid.NewString(),
			AppPath:    uuid.NewString(),
			ServerPath: uuid.NewString(),
		}
	}

	v2.WithAssumedFailedDeployment(t, tooManyVolumes)

	tooManyInitCommands := body.DeploymentCreate{Name: e2e.GenName()}
	tooManyInitCommands.InitCommands = make([]string, 10000)
	for i := range tooManyInitCommands.InitCommands {
		tooManyInitCommands.InitCommands[i] = uuid.NewString()
	}

	v2.WithAssumedFailedDeployment(t, tooManyInitCommands)
}

func TestCreateTooBig(t *testing.T) {
	t.Parallel()

	// Fetch the quota for the user
	quota := v2.GetUser(t, model.TestPowerUserID).Quota

	// Create a deployment that is too big CPU-wise
	cpuCores := quota.CpuCores / 2
	replicas := 3

	v2.WithAssumedFailedDeployment(t, body.DeploymentCreate{
		Name:     e2e.GenName(),
		Replicas: &replicas,
		CpuCores: &cpuCores,
	}, e2e.PowerUser)

	// Create a deployment that is too big RAM-wise
	ram := quota.RAM / 2
	replicas = 3

	v2.WithAssumedFailedDeployment(t, body.DeploymentCreate{
		Name:     e2e.GenName(),
		Replicas: &replicas,
		RAM:      &ram,
	}, e2e.PowerUser)
}

func TestCreateShared(t *testing.T) {
	t.Parallel()

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})
	team := v2.WithTeam(t, body.TeamCreate{
		Name:      e2e.GenName(),
		Resources: []string{deployment.ID},
		Members:   []body.TeamMemberCreate{{ID: model.TestDefaultUserID}},
	}, e2e.PowerUser)

	deploymentRead := v2.GetDeployment(t, deployment.ID)
	assert.Equal(t, []string{team.ID}, deploymentRead.Teams, "invalid teams on deployment")

	// Fetch team members deployments
	deployments := v2.ListDeployments(t, "?userId="+model.TestDefaultUserID, e2e.DefaultUser)
	assert.NotEmpty(t, deployments, "user has no deployments")

	hasDeployment := false
	for _, d := range deployments {
		if d.ID == deployment.ID {
			hasDeployment = true
		}
	}

	assert.True(t, hasDeployment, "deployment was not found in other user's deployments")
}

func TestCreateWithCustomSpecs(t *testing.T) {
	t.Parallel()

	cpuCores := 1.1
	ram := 1.1

	requestBody := body.DeploymentCreate{
		Name:     e2e.GenName(),
		CpuCores: &cpuCores,
		RAM:      &ram,
	}

	d, _ := v2.WithDeployment(t, requestBody)

	assert.Greater(t, d.Specs.CpuCores, 1.0, "cpu cores were not set")
	assert.Greater(t, d.Specs.RAM, 1.0, "ram was not set")
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	envValue := uuid.NewString()

	deploymentRead, _ := v2.WithDeployment(t, body.DeploymentCreate{
		Name:       e2e.GenName(),
		Visibility: model.VisibilityPublic,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
	})

	// Update deployment
	newVisibility := model.VisibilityPrivate
	deploymentUpdate := body.DeploymentUpdate{
		Envs: &[]body.Env{
			{
				Name:  e2e.GenName(),
				Value: uuid.NewString(),
			},
		},
		Visibility: &newVisibility,
		Volumes: &[]body.Volume{
			{
				Name:       "e2e-test",
				AppPath:    "/etc/test",
				ServerPath: "/",
			},
		},
	}

	v2.UpdateDeployment(t, deploymentRead.ID, deploymentUpdate)
}

func TestUpdateImage(t *testing.T) {
	t.Parallel()

	image1 := "nginx"
	image2 := "httpd"

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{
		Name:  e2e.GenName(),
		Image: &image1,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: "80",
			},
		},
	})

	deploymentUpdate := body.DeploymentUpdate{
		Image: &image2,
	}

	v2.UpdateDeployment(t, deployment.ID, deploymentUpdate)
}

func TestUpdateInternalPort(t *testing.T) {
	t.Parallel()

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	customPort := deployment.InternalPort + 1

	deploymentUpdate := body.DeploymentUpdate{
		Envs: &[]body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
	}

	v2.UpdateDeployment(t, deployment.ID, deploymentUpdate)
}

func TestUpdateName(t *testing.T) {
	t.Parallel()

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	newName := e2e.GenName()
	deploymentUpdate := body.DeploymentUpdate{
		Name: &newName,
	}

	v2.UpdateDeployment(t, deployment.ID, deploymentUpdate)
}

func TestCommand(t *testing.T) {
	t.Parallel()

	commands := []string{"restart"}

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	for _, command := range commands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, v2.DeploymentPath+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		time.Sleep(3 * time.Second)

		v2.WaitForDeploymentRunning(t, deployment.ID, func(deploymentRead *body.DeploymentRead) bool {
			//make sure it is accessible
			if deploymentRead.URL != nil {
				return v2.CheckUpURL(t, *deploymentRead.URL)
			}
			return false
		})
	}
}

func TestInvalidCommand(t *testing.T) {
	t.Parallel()

	invalidCommands := []string{"start", "stop"}

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	for _, command := range invalidCommands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, v2.DeploymentPath+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestFetchCiConfig(t *testing.T) {
	t.Parallel()

	image := "nginx"
	deploymentCustom, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})
	deploymentPrebuilt, _ := v2.WithDeployment(t, body.DeploymentCreate{
		Name:  e2e.GenName(),
		Image: &image,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: "80",
			},
		},
	})

	resp := e2e.DoGetRequest(t, v2.DeploymentPath+deploymentCustom.ID+"/ciConfig")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ciConfig body.CiConfig
	err := e2e.ReadResponseBody(t, resp, &ciConfig)
	assert.NoError(t, err, "ci config was not fetched")
	assert.NotEmpty(t, ciConfig.Config)

	// Not ci config for prebuilt deployments
	resp = e2e.DoGetRequest(t, v2.DeploymentPath+deploymentPrebuilt.ID+"/ciConfig")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestFetchLogs(t *testing.T) {
	t.Parallel()

	deployment, _ := v2.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	resp := e2e.DoGetRequest(t, v2.DeploymentPath+deployment.ID+"/logs-sse")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
}
