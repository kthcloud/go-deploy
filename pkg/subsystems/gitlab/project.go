package gitlab

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
)

func (client *Client) CreateProject(public *models.ProjectPublic) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gitlab project. details: %s", err)
	}

	project, _, err := client.GitLabClient.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:      gitlab.String(public.Name),
		ImportURL: &public.ImportURL,
	})

	if err != nil {
		return 0, makeError(err)
	}

	return project.ID, nil
}

func (client *Client) DeleteProject(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete gitlab project. details: %s", err)
	}

	_, err := client.GitLabClient.Projects.DeleteProject(id)

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) AttachCiFile(projectID int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach ci file to project. details: %s", err)
	}

	_, _, err := client.GitLabClient.RepositoryFiles.CreateFile(projectID, ".gitlab-ci.yml", &gitlab.CreateFileOptions{
		Branch:        gitlab.String("main"),
		Content:       gitlab.String("image: alpine:latest\n\nstages:\n  - deploy\n\ndeploy:\n  stage: deploy\n  script:\n    - echo \"Deploying\"\n  only:\n    - main\n"),
		AuthorEmail:   gitlab.String("deploy@cloud.cbh.kth.se"),
		AuthorName:    gitlab.String("go-deploy"),
		CommitMessage: gitlab.String("Add .gitlab-ci.yml"),
	}, nil)

	if err != nil {
		return makeError(err)
	}

	return nil
}
