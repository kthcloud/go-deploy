package gitlab

import (
	"fmt"
	"github.com/google/uuid"
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
