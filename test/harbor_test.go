package test

import (
	"errors"
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/harbor/models"
	"testing"
)

func TestEmptyProject(t *testing.T) {
	setup(t)

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.Identity,
		Password: conf.Env.Harbor.Secret,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.ProjectPublic{
		Name:   "acc-test",
		Public: false,
	}

	id, err := client.CreateProject(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = id

	if public.ID == 0 {
		t.Fatalf("no proxy host id received from client")
	}

	err = client.DeleteProject(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProject, err := client.ReadProject(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedProject != nil {
		t.Fatalf("failed to delete project")
	}
}

func TestUpdateProject(t *testing.T) {
	setup(t)

	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.Identity,
		Password: conf.Env.Harbor.Secret,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.ProjectPublic{
		Name:   "acc-test",
		Public: false,
	}

	id, err := client.CreateProject(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = id

	if public.ID == 0 {
		t.Fatalf("no id received from client")
	}

	public.Public = true

	err = client.UpdateProject(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	updatedProject, err := client.ReadProject(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if updatedProject.Public != true {
		t.Fatalf("failed to update project public=false -> public=true")
	}

	err = client.DeleteProject(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedProject, err := client.ReadProject(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedProject != nil {
		t.Fatalf("failed to delete project")
	}
}

func withProject() (*harbor.Client, *models.ProjectPublic, error) {
	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.Identity,
		Password: conf.Env.Harbor.Secret,
	})

	if err != nil {
		return nil, nil, err
	}

	public := &models.ProjectPublic{
		Name:   "acc-test",
		Public: false,
	}

	id, err := client.CreateProject(public)
	if err != nil {
		return nil, nil, err
	}

	public.ID = id

	if public.ID == 0 {
		return nil, nil, errors.New("no id received from client")
	}

	return client, public, nil
}

func deleteProject(public *models.ProjectPublic) error {
	client, err := harbor.New(&harbor.ClientConf{
		ApiUrl:   conf.Env.Harbor.Url,
		Username: conf.Env.Harbor.Identity,
		Password: conf.Env.Harbor.Secret,
	})
	if err != nil {
		return err
	}

	err = client.DeleteProject(public.ID)
	if err != nil {
		return err
	}

	deletedProject, err := client.ReadProject(public.ID)
	if err != nil {
		return err
	}

	if deletedProject != nil {
		return fmt.Errorf("failed to delete project")
	}

	return nil
}

func TestProjectWithRobot(t *testing.T) {
	setup(t)

	client, project, err := withProject()
	if err != nil {
		t.Fatalf(err.Error())
	}

	public := &models.RobotPublic{
		Name:        project.Name,
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Description: "some description",
		Disable:     false,
	}

	created, err := client.CreateRobot(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	public.ID = created.ID

	if public.ID == 0 {
		t.Fatalf("no id received from client")
	}

	createdRobot, err := client.ReadRobot(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if createdRobot.Disable != public.Disable {
		t.Fatalf("failed to create robot. field disable is invalid")
	}
	if createdRobot.Description != public.Description {
		t.Fatalf("failed to create robot. field description is invalid")
	}

	public.Disable = true
	public.Description = "another description"

	err = client.UpdateRobot(public)
	if err != nil {
		t.Fatalf(err.Error())
	}

	updatedRobot, err := client.ReadRobot(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if updatedRobot.Disable != public.Disable {
		t.Fatalf("failed to update robot disable status")
	}
	if updatedRobot.Description != public.Description {
		t.Fatalf("failed to update robot description")
	}

	err = client.DeleteRobot(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deletedRobot, err := client.ReadRobot(public.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if deletedRobot != nil {
		t.Fatalf("failed to delete robot")
	}

	err = deleteProject(project)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
