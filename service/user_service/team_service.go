package user_service

import (
	"errors"
	"fmt"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	userModel "go-deploy/models/sys/user"
	teamModel "go-deploy/models/sys/user/team"
	"go-deploy/service"
	"golang.org/x/exp/maps"
)

var TeamNameTakenErr = fmt.Errorf("team name taken")

func CreateTeam(id string, dtoCreateTeam *body.TeamCreate) (*teamModel.Team, error) {
	params := &teamModel.CreateParams{}
	params.FromDTO(dtoCreateTeam)

	teamClient := teamModel.New()
	team, err := teamClient.Create(id, params)
	if err != nil {
		if errors.Is(err, teamModel.NameTaken) {
			return nil, TeamNameTakenErr
		}
		return nil, err
	}

	return team, nil
}

func GetTeamByIdAuth(id string, auth *service.AuthInfo) (*teamModel.Team, error) {
	teamClient := teamModel.New()
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

func GetTeamListAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]teamModel.Team, error) {
	teamClient := teamModel.New()
	userClient := userModel.New()

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

func UpdateTeamAuth(id string, dtoUpdateTeam *body.TeamUpdate, auth *service.AuthInfo) (*teamModel.Team, error) {
	teamClient := teamModel.New()

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

	params := &teamModel.UpdateParams{}
	params.FromDTO(dtoUpdateTeam)

	err = teamClient.UpdateWithParamsByID(id, params)
	if err != nil {
		return nil, err
	}

	return teamClient.GetByID(id)
}

func DeleteTeamAuth(id string, auth *service.AuthInfo) error {
	teamClient := teamModel.New()

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
