package gitlab

import (
	"go-deploy/pkg/subsystems/gitlab/models"
	"go-deploy/test/acc"
	"testing"
)

func TestCreateProject(t *testing.T) {
	withDefaultProject(t)
}

func TestCreateImportProject(t *testing.T) {
	// A tiny public repo to test
	importURL := "https://github.com/rtyley/small-test-repo.git"

	p := &models.ProjectPublic{
		Name:      acc.GenName(),
		ImportURL: importURL,
	}

	withProject(t, p)
}
