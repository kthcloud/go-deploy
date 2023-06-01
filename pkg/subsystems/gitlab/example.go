package gitlab

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/xanzy/go-gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
)

func ExampleCreateProject() {
	client, err := New(&ClientConf{
		URL:   "https://gitlab.cloud.cbh.kth.se/api/v4/",
		Token: "glpat-FRhLxXJjBSsQzJbicJHj",
	})

	if err != nil {
		panic(err)
	}

	public := &models.ProjectPublic{
		Name:      "test" + "-" + uuid.NewString(),
		ImportURL: "https://github.com/saffronjam/go-deploy-placeholder",
	}

	id, err := client.CreateProject(public)

	if err != nil {
		panic(err)
	}

	fmt.Println(id)

	err = client.DeleteProject(id)
	if err != nil {
		panic(err)
	}
}

func ExampleListJobs() {
	client, err := New(&ClientConf{
		URL:   "https://gitlab.cloud.cbh.kth.se/api/v4/",
		Token: "glpat-FRhLxXJjBSsQzJbicJHj",
	})

	if err != nil {
		panic(err)
	}

	jobs, _, err := client.GitLabClient.Jobs.ListProjectJobs(16, &gitlab.ListJobsOptions{})
	if err != nil {
		panic(err)
	}

	for _, job := range jobs {
		var reader *bytes.Reader
		reader, _, err = client.GitLabClient.Jobs.GetTraceFile(16, job.ID, nil)
		if err != nil {
			continue
		}

		// read the content to a string
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(reader)
		newStr := buf.String()

		fmt.Println(newStr)
	}
}
