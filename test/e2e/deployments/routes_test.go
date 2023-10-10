package deployments

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go-deploy/models/dto/body"
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

func TestGetList(t *testing.T) {
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

func TestCreate(t *testing.T) {
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
		Name:    e2e.GenName("e2e"),
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

	_, _ = e2e.WithDeployment(t, requestBody)
}

func TestCreateWithCustomPort(t *testing.T) {
	// this test assumes that the default port is 8080
	customPort := 8081

	requestBody := body.DeploymentCreate{
		Name:    e2e.GenName("e2e"),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
		GitHub: nil,
		Zone:   nil,
	}

	_, _ = e2e.WithDeployment(t, requestBody)
}

func TestCreateWithCustomImage(t *testing.T) {
	// choose this setup so that it is reachable (pingable)
	customImage := "nginx:latest"
	customPort := 80

	requestBody := body.DeploymentCreate{
		Name:    e2e.GenName("e2e"),
		Private: false,
		Envs: []body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
		Image:  &customImage,
		GitHub: nil,
		Zone:   nil,
	}

	_, _ = e2e.WithDeployment(t, requestBody)
}

func TestCreateWithCustomDomain(t *testing.T) {
	customDomain := e2e.TestDomain

	requestBody := body.DeploymentCreate{
		Name:         e2e.GenName("e2e"),
		Private:      false,
		CustomDomain: &customDomain,
	}

	_, _ = e2e.WithDeployment(t, requestBody)
}

func TestCreateWithInvalidBody(t *testing.T) {
	longName := body.DeploymentCreate{Name: "e2e-"}
	for i := 0; i < 1000; i++ {
		longName.Name += uuid.NewString()
	}
	e2e.WithAssumedFailedDeployment(t, longName)

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
		e2e.WithAssumedFailedDeployment(t, requestBody)
	}

	tooManyEnvs := body.DeploymentCreate{Name: e2e.GenName("e2e")}
	tooManyEnvs.Envs = make([]body.Env, 10000)
	for i := range tooManyEnvs.Envs {
		tooManyEnvs.Envs[i] = body.Env{
			Name:  uuid.NewString(),
			Value: uuid.NewString(),
		}
	}

	e2e.WithAssumedFailedDeployment(t, tooManyEnvs)

	tooManyVolumes := body.DeploymentCreate{Name: e2e.GenName("e2e")}
	tooManyVolumes.Volumes = make([]body.Volume, 10000)
	for i := range tooManyVolumes.Volumes {
		tooManyVolumes.Volumes[i] = body.Volume{
			Name:       uuid.NewString(),
			AppPath:    uuid.NewString(),
			ServerPath: uuid.NewString(),
		}
	}

	e2e.WithAssumedFailedDeployment(t, tooManyVolumes)

	tooManyInitCommands := body.DeploymentCreate{Name: e2e.GenName("e2e")}
	tooManyInitCommands.InitCommands = make([]string, 10000)
	for i := range tooManyInitCommands.InitCommands {
		tooManyInitCommands.InitCommands[i] = uuid.NewString()
	}

	e2e.WithAssumedFailedDeployment(t, tooManyInitCommands)
}

func TestUpdate(t *testing.T) {
	envValue := uuid.NewString()

	deploymentRead, _ := e2e.WithDeployment(t, body.DeploymentCreate{
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

	e2e.WaitForJobFinished(t, deploymentUpdated.JobID, func(jobRead *body.JobRead) bool {
		return true
	})

	e2e.WaitForDeploymentRunning(t, deploymentRead.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return e2e.CheckUpURL(t, *deploymentRead.URL)
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

func TestUpdateImage(t *testing.T) {
	image1 := "nginx"
	image2 := "httpd"

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name:  e2e.GenName("e2e"),
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

	resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID, deploymentUpdate)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deploymentUpdated body.DeploymentUpdated
	err := e2e.ReadResponseBody(t, resp, &deploymentUpdated)
	assert.NoError(t, err, "deployment was not updated")

	e2e.WaitForJobFinished(t, deploymentUpdated.JobID, nil)

	// check if the deployment was updated
	resp = e2e.DoGetRequest(t, "/deployments/"+deploymentUpdated.ID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deploymentReadUpdated body.DeploymentRead
	err = e2e.ReadResponseBody(t, resp, &deploymentReadUpdated)
	assert.NoError(t, err, "deployment was not updated")

	assert.Equal(t, image2, *deploymentReadUpdated.Image)
}

func TestUpdateInternalPort(t *testing.T) {
	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	customPort := deployment.InternalPort + 1

	deploymentUpdate := body.DeploymentUpdate{
		Envs: &[]body.Env{
			{
				Name:  "PORT",
				Value: strconv.Itoa(customPort),
			},
		},
	}

	resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID, deploymentUpdate)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deploymentUpdated body.DeploymentUpdated
	err := e2e.ReadResponseBody(t, resp, &deploymentUpdated)
	assert.NoError(t, err, "deployment was not updated")

	e2e.WaitForJobFinished(t, deploymentUpdated.JobID, nil)
	e2e.WaitForDeploymentRunning(t, deployment.ID, func(deploymentRead *body.DeploymentRead) bool {
		//make sure it is accessible
		if deploymentRead.URL != nil {
			return e2e.CheckUpURL(t, *deploymentRead.URL)
		}
		return false
	})

	// check if the deployment was updated
	resp = e2e.DoGetRequest(t, "/deployments/"+deployment.ID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deploymentReadUpdated body.DeploymentRead
	err = e2e.ReadResponseBody(t, resp, &deploymentReadUpdated)
	assert.NoError(t, err, "deployment was not created")

	assert.Equal(t, customPort, deploymentReadUpdated.InternalPort)
	assert.True(t, e2e.CheckUpURL(t, *deploymentReadUpdated.URL))
}

func TestCommand(t *testing.T) {
	commands := []string{"restart"}

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

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
	invalidCommands := []string{"start", "stop"}

	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

	for _, command := range invalidCommands {
		reqBody := body.DeploymentCommand{Command: command}
		resp := e2e.DoPostRequest(t, "/deployments/"+deployment.ID+"/command", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestFetchCiConfig(t *testing.T) {
	image := "nginx"
	deploymentCustom, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})
	deploymentPrebuilt, _ := e2e.WithDeployment(t, body.DeploymentCreate{
		Name:  e2e.GenName("e2e"),
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

	// not ci config for prebuilt deployments
	resp = e2e.DoGetRequest(t, "/deployments/"+deploymentPrebuilt.ID+"/ciConfig")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TODO: Update to SSE once it works
func TestFetchLogs(t *testing.T) {
	deployment, _ := e2e.WithDeployment(t, body.DeploymentCreate{Name: e2e.GenName("e2e")})

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

	_, _ = e2e.WithDeployment(t, requestBody)

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

	e2e.WaitForStorageManagerRunning(t, storageManager.ID, func(storageManagerRead *body.StorageManagerRead) bool {
		//make sure it is accessible
		if storageManager.URL != nil {
			return e2e.CheckUpURL(t, *storageManager.URL)
		}
		return false
	})
}
