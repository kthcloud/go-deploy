package user_service

import (
	"errors"
	"fmt"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	deploymentModel "go-deploy/models/sys/deployment"
	userModels "go-deploy/models/sys/user"
	teamModels "go-deploy/models/sys/user/team"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service"
	"go-deploy/utils"
	"golang.org/x/exp/maps"
	"time"
)

var TeamNameTakenErr = fmt.Errorf("team name taken")
var TeamNotFoundErr = fmt.Errorf("team not found")

func CreateTeam(id, ownerID string, dtoCreateTeam *body.TeamCreate, auth *service.AuthInfo) (*teamModels.Team, error) {
	params := &teamModels.CreateParams{}
	params.FromDTO(dtoCreateTeam, func(resourceID string) *teamModels.Resource {
		return getResourceIfAccessible(ownerID, resourceID, auth)
	})

	teamClient := teamModels.New()
	team, err := teamClient.Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, teamModels.NameTakenErr) {
			return nil, TeamNameTakenErr
		}
		return nil, err
	}

	return team, nil
}

func GetTeamByIdAuth(id string, auth *service.AuthInfo) (*teamModels.Team, error) {
	teamClient := teamModels.New()
	team, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	if !auth.IsAdmin && !team.HasMember(auth.UserID) {
		return nil, nil
	}

	return team, nil
}

func GetTeamListAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]teamModels.Team, error) {
	teamClient := teamModels.New()
	userClient := userModels.New()

	if pagination != nil {
		teamClient.AddPagination(pagination.Page, pagination.PageSize)
		userClient.AddPagination(pagination.Page, pagination.PageSize)
	}

	var withUserID *string
	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		withUserID = userID
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		withUserID = &auth.UserID
	}

	if withUserID != nil {
		user, err := userClient.GetByID(*withUserID)
		if err != nil {
			return nil, err
		}

		if user == nil {
			return nil, nil
		}

		teams, err := user.GetTeamMap()
		if err != nil {
			return nil, err
		}

		return maps.Values(teams), nil
	}

	return teamClient.GetAll()
}

func UpdateTeamAuth(id string, dtoUpdateTeam *body.TeamUpdate, auth *service.AuthInfo) (*teamModels.Team, error) {
	teamClient := teamModels.New()

	team, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, TeamNotFoundErr
	}

	if team.OwnerID != auth.UserID && !auth.IsAdmin && !team.HasMember(auth.UserID) {
		return nil, nil
	}

	params := &teamModels.UpdateParams{}
	params.FromDTO(dtoUpdateTeam, func(resourceID string) *teamModels.Resource {
		return getResourceIfAccessible(team.OwnerID, resourceID, auth)
	})

	err = teamClient.UpdateWithParamsByID(id, params)
	if err != nil {
		return nil, err
	}

	afterUpdate, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	if afterUpdate == nil {
		return nil, TeamNotFoundErr
	}

	return afterUpdate, nil
}

func DeleteTeamAuth(id string, auth *service.AuthInfo) error {
	teamClient := teamModels.New()

	team, err := teamClient.GetByID(id)
	if err != nil {
		return err
	}

	if team == nil {
		return nil
	}

	if !auth.IsAdmin && !team.HasMember(auth.UserID) {
		return nil
	}

	return teamClient.DeleteByID(id)
}

func getResourceIfAccessible(userID string, resourceID string, auth *service.AuthInfo) *teamModels.Resource {
	// try to fetch deployment
	dClient := deploymentModel.New()
	vClient := vmModel.New()

	if !auth.IsAdmin {
		dClient.RestrictToUser(userID)
		vClient.RestrictToUser(userID)
	}

	isOwner, err := dClient.ExistsByID(resourceID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to fetch deployment when checking user access when creating team: %w", err))
		return nil
	}

	if isOwner {
		return &teamModels.Resource{
			ID:      resourceID,
			Type:    teamModels.ResourceTypeDeployment,
			AddedAt: time.Now(),
		}
	}

	// try to fetch vm
	isOwner, err = vmModel.New().RestrictToUser(userID).ExistsByID(resourceID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to fetch vm when checking user access when creating team: %w", err))
		return nil
	}

	if isOwner {
		return &teamModels.Resource{
			ID:      resourceID,
			Type:    teamModels.ResourceTypeVM,
			AddedAt: time.Now(),
		}
	}

	return nil
}
