package deployments

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
	"go-deploy/test/e2e"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	e2e.Setup()
	code := m.Run()
	e2e.Shutdown()
	os.Exit(code)
}

func TestGetDeployments(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/deployments")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deployments []body.DeploymentRead
	err := e2e.ReadResponseBody(t, resp, &deployments)
	assert.NoError(t, err, "deployments were not fetched")

	for _, deployment := range deployments {
		assert.NotEmpty(t, deployment.ID, "deployment id was empty")
		assert.NotEmpty(t, deployment.Name, "deployment name was empty")
	}
}

func TestGetStorageManagers(t *testing.T) {
	resp := e2e.DoGetRequest(t, "/storageManagers")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var storageManagers []body.StorageManagerRead
	err := e2e.ReadResponseBody(t, resp, &storageManagers)
	assert.NoError(t, err, "storage managers were not fetched")

	for _, storageManager := range storageManagers {
		assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
		assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")
		assert.NotEmpty(t, storageManager.URL, "storage manager url was empty")
	}
}

func TestCreateDeployment(t *testing.T) {
	// in order to test with GitHub, you need to set the following env variables:
	var token string
	var repositoryID int64

	var github *body.GitHub
	if token != "" && repositoryID != 0 {
		github = &body.GitHub{
			Token:        token,
			RepositoryID: repositoryID,
		}
	}

	envValue := uuid.NewString()

	requestBody := body.DeploymentCreate{
		Name:    "e2e-" + strings.ReplaceAll(uuid.NewString()[:10], "-", ""),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
		GitHub: github,
		Zone:   nil,
	}

	_ = withDeployment(t, requestBody)
}

func TestCreateDeploymentWithCustomPort(t *testing.T) {
	envValue := uuid.NewString()

	// this test assumes that the default port is 8080
	customPort := 8081

	requestBody := body.DeploymentCreate{
		Name:    "e2e-" + strings.ReplaceAll(uuid.NewString()[:10], "-", ""),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
		GitHub:       nil,
		Zone:         nil,
		InternalPort: &customPort,
	}

	_ = withDeployment(t, requestBody)
}

func TestCreateDeploymentWithInvalidBody(t *testing.T) {
	longName := body.DeploymentCreate{Name: "e2e-"}
	for i := 0; i < 1000; i++ {
		longName.Name += uuid.NewString()
	}
	withAssumedFailedDeployment(t, longName)

	invalidNames := []string{
		e2e.GenName("e2e") + "-",
		e2e.GenName("e2e") + "- ",
		e2e.GenName("e2e") + ".",
		"." + e2e.GenName("e2e"),
		e2e.GenName("e2e") + " " + e2e.GenName("e2e"),
		e2e.GenName("e2e") + "%",
		e2e.GenName("e2e") + "!",
		e2e.GenName("e2e") + "%" + e2e.GenName("e2e"),
	}

	for _, name := range invalidNames {
		requestBody := body.DeploymentCreate{Name: name}
		withAssumedFailedDeployment(t, requestBody)
	}

	tooManyEnvs := body.DeploymentCreate{Name: e2e.GenName("e2e")}
	tooManyEnvs.Envs = make([]body.Env, 10000)
	for i := range tooManyEnvs.Envs {
		tooManyEnvs.Envs[i] = body.Env{
			Name:  uuid.NewString(),
			Value: uuid.NewString(),
		}
	}

	withAssumedFailedDeployment(t, tooManyEnvs)

	tooManyVolumes := body.DeploymentCreate{Name: e2e.GenName("e2e")}
	tooManyVolumes.Volumes = make([]body.Volume, 10000)
	for i := range tooManyVolumes.Volumes {
		tooManyVolumes.Volumes[i] = body.Volume{
			Name:       uuid.NewString(),
			AppPath:    uuid.NewString(),
			ServerPath: uuid.NewString(),
		}
	}

	withAssumedFailedDeployment(t, tooManyVolumes)

	tooManyInitCommands := body.DeploymentCreate{Name: e2e.GenName("e2e")}
	tooManyInitCommands.InitCommands = make([]string, 10000)
	for i := range tooManyInitCommands.InitCommands {
		tooManyInitCommands.InitCommands[i] = uuid.NewString()
	}

	withAssumedFailedDeployment(t, tooManyInitCommands)
}

func TestUpdateDeployment(t *testing.T) {
	envValue := uuid.NewString()

	deploymentRead := withDeployment(t, body.DeploymentCreate{
		Name:    e2e.GenName("e2e"),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
		GitHub: nil,
		Zone:   nil,
	})

	// update deployment
	newEnvValue := uuid.NewString()
	newPrivateValue := !deploymentRead.Private
	deploymentUpdate := body.DeploymentUpdate{
		Envs: &[]body.Env{
			{
				Name:  "e2e",
				Value: newEnvValue,
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

	resp := e2e.DoPostRequest(t, "/deployments/"+deploymentRead.ID, deploymentUpdate)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deploymentUpdated body.DeploymentUpdated
	err := e2e.ReadResponseBody(t, resp, &deploymentUpdated)
	assert.NoError(t, err, "deployment was not updated")

	waitForJobFinished(t, deploymentUpdated.JobID, func(jobRead *body.JobRead) bool {
		return true
	})

	waitForDeploymentRunning(t, deploymentRead.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return checkUpURL(t, *deploymentRead.URL)
		}

		if deploymentRead.Private {
			return true
		}
		return false
	})

	// check if the deployment was updated
	resp = e2e.DoGetRequest(t, "/deployments/"+deploymentRead.ID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deploymentReadUpdated body.DeploymentRead
	err = e2e.ReadResponseBody(t, resp, &deploymentReadUpdated)
	assert.NoError(t, err, "deployment was not created")

	assert.Equal(t, newEnvValue, deploymentReadUpdated.Envs[0].Value)
	assert.Equal(t, newPrivateValue, deploymentReadUpdated.Private)
	assert.NotEmpty(t, deploymentReadUpdated.Volumes)
}

func TestDeploymentCommand(t *testing.T) {
	commands := []string{"restart"}

	deployment := withDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	for _, command := range commands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		time.Sleep(3 * time.Second)

		waitForDeploymentRunning(t, deployment.ID, func(deploymentRead *body.DeploymentRead) bool {
			//make sure it is accessible
			if deploymentRead.URL != nil {
				return checkUpURL(t, *deploymentRead.URL)
			}
			return false
		})
	}
}

func TestDeploymentInvalidCommand(t *testing.T) {
	invalidCommands := []string{"start", "stop"}

	deployment := withDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	for _, command := range invalidCommands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestFetchDeploymentCiConfig(t *testing.T) {
	deployment := withDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	resp := e2e.DoGetRequest(t, "/deployments/"+deployment.ID+"/ciConfig")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ciConfig body.CiConfig
	err := e2e.ReadResponseBody(t, resp, &ciConfig)
	assert.NoError(t, err, "ci config was not fetched")
	assert.NotEmpty(t, ciConfig.Config)
}

func TestFetchDeploymentLogs(t *testing.T) {
	deployment := withDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	url := e2e.CreateServerUrlWithProtocol("ws", "/deployments/"+deployment.ID+"/logs")

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err, "websocket connection was not established")

	err = conn.WriteMessage(websocket.TextMessage, []byte("Bearer test user"))
	assert.NoError(t, err, "websocket bearer token message was not sent")

	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	assert.NoError(t, err, "websocket read deadline was not set")

	_, message, err := conn.ReadMessage()
	assert.NoError(t, err, "websocket message was not read")
	assert.NotEmpty(t, message)

	err = conn.Close()
	assert.NoError(t, err, "websocket connection was not closed")
}

func TestCreateStorageManager(t *testing.T) {
	// in order to test this, we need to create a deployment with volumes

	envValue := uuid.NewString()

	requestBody := body.DeploymentCreate{
		Name:    "e2e-" + strings.ReplaceAll(uuid.NewString()[:10], "-", ""),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "e2e",
				Value: envValue,
			},
		},
		GitHub: nil,
		Zone:   nil,
		Volumes: []body.Volume{
			{
				Name:    "e2e-test",
				AppPath: "/etc/test",
				// keeping this at root ensures that deployment can start,
				// since we don't have control of the folders
				ServerPath: "/",
			},
		},
	}

	_ = withDeployment(t, requestBody)

	// make sure the storage manager has to be created
	time.Sleep(30 * time.Second)

	// now the storage manager should be available
	// TODO: update this part of the test when storage manager id is exposed in the deployment/user
	var storageManager body.StorageManagerRead
	{
		resp := e2e.DoGetRequest(t, "/storageManagers")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var storageManagers []body.StorageManagerRead
		err := e2e.ReadResponseBody(t, resp, &storageManagers)
		assert.NoError(t, err, "storage managers were not fetched")

		idx := -1
		for i, sm := range storageManagers {
			if sm.OwnerID == e2e.TestUserID {
				idx = i
				break
			}
		}

		assert.NotEqual(t, -1, idx, "storage manager was not found")

		storageManager = storageManagers[idx]
	}

	assert.NotEmpty(t, storageManager.ID, "storage manager id was empty")
	assert.NotEmpty(t, storageManager.OwnerID, "storage manager owner id was empty")
	assert.NotEmpty(t, storageManager.URL, "storage manager url was empty")

	waitForStorageManagerRunning(t, storageManager.ID, func(storageManagerRead *body.StorageManagerRead) bool {
		//make sure it is accessible
		if storageManager.URL != nil {
			return checkUpURL(t, *storageManager.URL)
		}
		return false
	})
}
