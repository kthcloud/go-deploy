package deployments

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/v1/body"
	"go-deploy/test/e2e"
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
		"?userId=" + e2e.PowerUserID + "&page=1&pageSize=3",
		"?userId=" + e2e.DefaultUserID + "&page=1&pageSize=3",
	}

	for _, query := range queries {
		e2e.ListDeployments(t, query)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	requestBody := body.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
		Envs: []body.Env{
			{
				Name:  e2e.GenName(),
				Value: uuid.NewString(),
			},
		},
	}

	e2e.WithDeployment(t, requestBody)
}

func TestCreateWithCustomPort(t *testing.T) {
	t.Parallel()

	// This test assumes that the default port is 8080
	customPort := 8081

	requestBody := body.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
	}

	e2e.WithDeployment(t, requestBody)
}

func TestCreateWithCustomImage(t *testing.T) {
	t.Parallel()

	// This setup is chosen to make the deployment reachable (pingable)
	customImage := "nginx:latest"
	customPort := 80

	requestBody := body.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
		Image: &customImage,
	}

	e2e.WithDeployment(t, requestBody)
}

func TestCreateWithCustomDomain(t *testing.T) {
	t.Parallel()

	customDomain := e2e.TestDomain

	requestBody := body.DeploymentCreate{
		Name:         e2e.GenName(),
		Private:      false,
		CustomDomain: &customDomain,
	}

	e2e.WithDeployment(t, requestBody)
}

func TestCreateWithInvalidBody(t *testing.T) {
	t.Parallel()

	longName := body.DeploymentCreate{Name: "e2e-"}
	for i := 0; i < 1000; i++ {
		longName.Name += uuid.NewString()
	}
	e2e.WithAssumedFailedDeployment(t, longName)

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
		e2e.WithAssumedFailedDeployment(t, requestBody)
	}

	tooManyEnvs := body.DeploymentCreate{Name: e2e.GenName()}
	tooManyEnvs.Envs = make([]body.Env, 10000)
	for i := range tooManyEnvs.Envs {
		tooManyEnvs.Envs[i] = body.Env{
			Name:  uuid.NewString(),
			Value: uuid.NewString(),
		}
	}

	e2e.WithAssumedFailedDeployment(t, tooManyEnvs)

	tooManyVolumes := body.DeploymentCreate{Name: e2e.GenName()}
	tooManyVolumes.Volumes = make([]body.Volume, 10000)
	for i := range tooManyVolumes.Volumes {
		tooManyVolumes.Volumes[i] = body.Volume{
			Name:       uuid.NewString(),
			AppPath:    uuid.NewString(),
			ServerPath: uuid.NewString(),
		}
	}

	e2e.WithAssumedFailedDeployment(t, tooManyVolumes)

	tooManyInitCommands := body.DeploymentCreate{Name: e2e.GenName()}
	tooManyInitCommands.InitCommands = make([]string, 10000)
	for i := range tooManyInitCommands.InitCommands {
		tooManyInitCommands.InitCommands[i] = uuid.NewString()
	}

	e2e.WithAssumedFailedDeployment(t, tooManyInitCommands)
}

func TestCreateShared(t *testing.T) {
	t.Parallel()

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})
	team := e2e.WithTeam(t, body.TeamCreate{
		Name:      e2e.GenName(),
		Resources: []string{deployment.ID},
		Members:   []body.TeamMemberCreate{{ID: e2e.PowerUserID}},
	})

	deploymentRead := e2e.GetDeployment(t, deployment.ID)
	assert.Equal(t, []string{team.ID}, deploymentRead.Teams, "invalid teams on deployment")

	// Fetch team members deployments
	deployments := e2e.ListDeployments(t, "?userId="+e2e.PowerUserID)
	assert.NotEmpty(t, deployments, "user has no deployments")

	hasDeployment := false
	for _, d := range deployments {
		if d.ID == deployment.ID {
			hasDeployment = true
		}
	}

	assert.True(t, hasDeployment, "deployment was not found in other user's deployments")
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	envValue := uuid.NewString()

	deploymentRead, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name:    e2e.GenName(),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
	})

	// Update deployment
	newPrivateValue := !deploymentRead.Private
	deploymentUpdate := body.DeploymentUpdate{
		Envs: &[]body.Env{
			{
				Name:  e2e.GenName(),
				Value: uuid.NewString(),
			},
		},
		Private: &newPrivateValue,
		Volumes: &[]body.Volume{
			{
				Name:       "e2e-test",
				AppPath:    "/etc/test",
				ServerPath: "/",
			},
		},
	}

	e2e.UpdateDeployment(t, deploymentRead.ID, deploymentUpdate)
}

func TestUpdateImage(t *testing.T) {
	t.Parallel()

	image1 := "nginx"
	image2 := "httpd"

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{
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

	e2e.UpdateDeployment(t, deployment.ID, deploymentUpdate)
}

func TestUpdateInternalPort(t *testing.T) {
	t.Parallel()

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	customPort := deployment.InternalPort + 1

	deploymentUpdate := body.DeploymentUpdate{
		Envs: &[]body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
	}

	e2e.UpdateDeployment(t, deployment.ID, deploymentUpdate)
}

func TestCommand(t *testing.T) {
	t.Parallel()

	commands := []string{"restart"}

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	for _, command := range commands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		time.Sleep(3 * time.Second)

		e2e.WaitForDeploymentRunning(t, deployment.ID, func(deploymentRead *body.DeploymentRead) bool {
			//make sure it is accessible
			if deploymentRead.URL != nil {
				return e2e.CheckUpURL(t, *deploymentRead.URL)
			}
			return false
		})
	}
}

func TestInvalidCommand(t *testing.T) {
	t.Parallel()

	invalidCommands := []string{"start", "stop"}

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	for _, command := range invalidCommands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestFetchCiConfig(t *testing.T) {
	t.Parallel()

	image := "nginx"
	deploymentCustom, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})
	deploymentPrebuilt, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name:  e2e.GenName(),
		Image: &image,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: "80",
			},
		},
	})

	resp := e2e.DoGetRequest(t, "/deployments/"+deploymentCustom.ID+"/ciConfig")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ciConfig body.CiConfig
	err := e2e.ReadResponseBody(t, resp, &ciConfig)
	assert.NoError(t, err, "ci config was not fetched")
	assert.NotEmpty(t, ciConfig.Config)

	// Not ci config for prebuilt deployments
	resp = e2e.DoGetRequest(t, "/deployments/"+deploymentPrebuilt.ID+"/ciConfig")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestFetchLogs(t *testing.T) {
	t.Parallel()

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName()})

	resp := e2e.DoGetRequest(t, "/deployments/"+deployment.ID+"/logs-sse")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
}
