package gitlab

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"go-deploy/models/sys/deployment"
	"go-deploy/pkg/subsystems/gitlab/models"
	"gopkg.in/yaml.v3"
	"strings"
)

func (client *Client) ReadProject(id int) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gitlab project. details: %w", err)
	}

	project, resp, err := client.GitLabClient.Projects.GetProject(id, nil)
	if resp.StatusCode == 404 {
		return nil, nil
	}
	
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateProjectPublicFromGet(project), nil
}

func (client *Client) CreateProject(public *models.ProjectPublic) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gitlab project. details: %w", err)
	}

	project, _, err := client.GitLabClient.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:      gitlab.String(public.Name),
		ImportURL: &public.ImportURL,
	})

	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateProjectPublicFromGet(project), nil
}

func (client *Client) DeleteProject(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete gitlab project. details: %w", err)
	}

	_, err := client.GitLabClient.Projects.DeleteProject(id)
	if err != nil {
		if strings.Contains(err.Error(), "Project Not Found") {
			return nil
		}
		return makeError(err)
	}

	return nil
}

func (client *Client) AttachCiFile(projectID int, branch string, content deployment.GitLabCiConfig) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach ci file to project. details: %w", err)
	}

	contentBytes, err := yaml.Marshal(content)
	if err != nil {
		return makeError(err)
	}

	contentString := string(contentBytes)

	_, _, err = client.GitLabClient.RepositoryFiles.CreateFile(projectID, ".gitlab-ci.yml", &gitlab.CreateFileOptions{
		Branch:        gitlab.String(branch),
		Content:       gitlab.String(contentString),
		AuthorEmail:   gitlab.String("deploy@cloud.cbh.kth.se"),
		AuthorName:    gitlab.String("deploy"),
		CommitMessage: gitlab.String("Add .gitlab-ci.yml"),
	}, nil)

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) GetJobs(projectID int) ([]*gitlab.Job, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get job for project %s. details: %w", projectID, err)
	}

	jobs, _, err := client.GitLabClient.Jobs.ListProjectJobs(projectID, nil)
	if err != nil {
		return nil, makeError(err)
	}

	return jobs, nil
}
