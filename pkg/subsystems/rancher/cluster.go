package rancher

import (
	"encoding/json"
	"fmt"
	"go-deploy/pkg/subsystems/rancher/models"
	"io"
	"net/http"
	"strings"
)

func (c *Client) ReadCluster(id string) (*models.ClusterPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read cluster. details: %w", err)
	}

	cluster, err := c.RancherClient.Cluster.ById(id)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, nil
		}

		return nil, makeError(err)
	}

	return models.CreateClusterPublicFromRead(cluster), nil
}

func (c *Client) ReadClusterKubeConfig(id string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read cluster kubeconfig. details: %w", err)
	}

	fullUrl := c.URL + "/clusters/" + id + "?action=generateKubeconfig"
	req, err := http.NewRequest("POST", fullUrl, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return "", nil
		}

		return "", makeError(err)
	}

	req.SetBasicAuth(c.ApiKey, c.Secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", makeError(err)
	}

	if resp.StatusCode == 404 {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", makeError(fmt.Errorf("failed to read cluster kubeconfig. status code: %d", resp.StatusCode))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", makeError(err)
	}

	var kubeConfig models.RancherKubeConfig
	err = json.Unmarshal(content, &kubeConfig)
	if err != nil {
		return "", makeError(err)
	}

	return kubeConfig.Config, nil
}
