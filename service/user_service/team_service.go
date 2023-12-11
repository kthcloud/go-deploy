package user_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	deploymentModel "go-deploy/models/sys/deployment"
	notificationModel "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	userModels "go-deploy/models/sys/user"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service"
	"go-deploy/service/notification_service"
	"go-deploy/utils"
	"golang.org/x/exp/maps"
	"time"
)

var TeamNameTakenErr = fmt.Errorf("team name taken")
var TeamNotFoundErr = fmt.Errorf("team not found")
var BadInviteCodeErr = fmt.Errorf("bad invite code")
var NotInvitedErr = fmt.Errorf("not invited")

func CreateTeam(id, ownerID string, dtoCreateTeam *body.TeamCreate, auth *service.AuthInfo) (*teamModels.Team, error) {
	params := &teamModels.CreateParams{}
	params.FromDTO(dtoCreateTeam,
		func(resourceID string) *teamModels.Resource { return getResourceIfAccessible(resourceID, auth) },
		func(memberDTO *body.TeamMemberCreate) *teamModels.Member {
			member := &teamModels.Member{
				ID:       memberDTO.ID,
				TeamRole: memberDTO.TeamRole,
				AddedAt:  time.Now(),
			}

			if auth.IsAdmin {
				member.MemberStatus = teamModels.MemberStatusJoined
				member.JoinedAt = time.Now()
			} else {
				member.MemberStatus = teamModels.MemberStatusInvited
				member.InvitationCode = createInvitationCode()
			}

			return member
		},
	)

	team, err := teamModels.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, teamModels.NameTakenErr) {
			return nil, TeamNameTakenErr
		}
		return nil, err
	}

	// send invitations to every member that received an invitation code
	for _, member := range params.MemberMap {
		if member.InvitationCode != "" {
			err = createInvitation(member.ID, team.ID, team.Name, member.InvitationCode)
			if err != nil {
				return nil, err
			}
		}
	}

	return team, nil
}

func JoinTeam(id string, dtoTeamJoin *body.TeamJoin, auth *service.AuthInfo) (*teamModels.Team, error) {
	params := &teamModels.JoinParams{}
	params.FromDTO(dtoTeamJoin)

	teamClient := teamModels.New()
	team, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, TeamNotFoundErr
	}

	if team.GetMemberMap()[auth.UserID].MemberStatus != teamModels.MemberStatusInvited {
		return team, NotInvitedErr
	}

	if team.GetMemberMap()[auth.UserID].InvitationCode != params.InvitationCode {
		return nil, BadInviteCodeErr
	}

	updatedMember := team.GetMemberMap()[auth.UserID]
	updatedMember.MemberStatus = teamModels.MemberStatusJoined
	updatedMember.JoinedAt = time.Now()

	err = teamClient.UpdateMember(id, auth.UserID, &updatedMember)
	if err != nil {
		return nil, err
	}

	return teamClient.GetByID(id)
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

func ListTeamsAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]teamModels.Team, error) {
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

	return teamClient.List()
}

func UpdateTeamAuth(id string, dtoUpdateTeam *body.TeamUpdate, auth *service.AuthInfo) (*teamModels.Team, error) {
	teamClient := teamModels.New()

	team, err := teamClient.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	if team.OwnerID != auth.UserID && !auth.IsAdmin && !team.HasMember(auth.UserID) {
		return nil, nil
	}

	params := &teamModels.UpdateParams{}
	params.FromDTO(dtoUpdateTeam,
		func(resourceID string) *teamModels.Resource { return getResourceIfAccessible(resourceID, auth) },
		func(memberDTO *body.TeamMemberUpdate) *teamModels.Member {
			if existing := team.GetMember(memberDTO.ID); existing != nil {
				existing.TeamRole = teamModels.MemberRoleAdmin
				return existing
			}

			member := &teamModels.Member{
				ID:       memberDTO.ID,
				TeamRole: teamModels.MemberRoleAdmin,
				AddedAt:  time.Now(),
			}

			if auth.IsAdmin {
				member.MemberStatus = teamModels.MemberStatusJoined
				member.JoinedAt = time.Now()
			} else {
				member.MemberStatus = teamModels.MemberStatusInvited
				member.InvitationCode = createInvitationCode()
			}

			return member
		},
	)

	// if new user, set timestamp
	if params.MemberMap != nil {
		for _, member := range *params.MemberMap {
			// don't invite users that are already joined
			if existing := team.GetMember(member.ID); existing != nil && existing.MemberStatus == teamModels.MemberStatusJoined {
				continue
			}

			if member.InvitationCode != "" {
				err = createInvitation(member.ID, team.ID, team.Name, member.InvitationCode)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// if new resource, set timestamp
	if params.ResourceMap != nil {
		for _, resource := range *params.ResourceMap {
			if resourceDB, ok := team.GetResourceMap()[resource.ID]; ok {
				resource.AddedAt = resourceDB.AddedAt
			} else {
				resource.AddedAt = time.Now()
			}
			(*params.ResourceMap)[resource.ID] = resource
		}
	}

	err = teamClient.UpdateWithParams(id, params)
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

	err = teamClient.MarkUpdated(id)
	if err != nil {
		return nil, err
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

	if !auth.IsAdmin && team.OwnerID != auth.UserID {
		return nil
	}

	return teamClient.DeleteByID(id)
}

func getResourceIfAccessible(resourceID string, auth *service.AuthInfo) *teamModels.Resource {
	// try to fetch deployment
	dClient := deploymentModel.New()
	vClient := vmModel.New()

	if !auth.IsAdmin {
		dClient.RestrictToOwner(auth.UserID)
		vClient.RestrictToOwner(auth.UserID)
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
	isOwner, err = vmModel.New().ExistsByID(resourceID)
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

func createInvitation(userID, teamID, teamName, invitationCode string) error {
	return notification_service.CreateNotification(uuid.NewString(), userID, &notificationModel.CreateParams{
		Type: notificationModel.TypeTeamInvite,
		Content: map[string]interface{}{
			"id":   teamID,
			"name": teamName,
			"code": invitationCode,
		},
	})
}

func createInvitationCode() string {
	return utils.HashString(uuid.NewString())
}
