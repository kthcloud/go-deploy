package teams

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/notification_repo"
	"go-deploy/pkg/db/resources/team_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	sErrors "go-deploy/service/errors"
	utils2 "go-deploy/service/utils"
	"go-deploy/service/v2/teams/opts"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Get gets a team by ID
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.Team, error) {
	_ = utils2.GetFirstOrDefault(opts)

	tmc := team_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		tmc.WithUserID(c.V2.Auth().User.ID)
	}

	return c.Team(id, tmc)
}

// List lists teams
func (c *Client) List(opts ...opts.ListOpts) ([]model.Team, error) {
	o := utils2.GetFirstOrDefault(opts)

	tmc := team_repo.New()

	if o.Pagination != nil {
		tmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != "" {
		// Specific user's teams are requested
		if !c.V2.HasAuth() || c.V2.Auth().User.ID == o.UserID || c.V2.Auth().User.IsAdmin {
			effectiveUserID = o.UserID
		} else {
			// User cannot access the other user's resources
			return nil, nil
		}
	} else {
		// All teams are requested
		if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
			effectiveUserID = c.V2.Auth().User.ID
		}
	}

	if effectiveUserID != "" {
		tmc.WithUserID(effectiveUserID)
	}

	if o.ResourceID != "" {
		tmc.WithResourceID(o.ResourceID)
	}

	return c.Teams(tmc)
}

// ListIDs lists team IDs
//
// This is a more lightweight version of List when only the IDs are needed
// TODO: Fetch only IDs, right now ListIDs == List
func (c *Client) ListIDs(opts ...opts.ListOpts) ([]string, error) {
	o := utils2.GetFirstOrDefault(opts)

	tmc := team_repo.New()

	if o.Pagination != nil {
		tmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != "" {
		// Specific user's teams are requested
		if !c.V2.HasAuth() || c.V2.Auth().User.ID == o.UserID || c.V2.Auth().User.IsAdmin {
			effectiveUserID = o.UserID
		} else {
			// User cannot access the other user's resources
			return nil, nil
		}
	} else {
		// All teams are requested
		if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
			effectiveUserID = c.V2.Auth().User.ID
		}
	}

	if effectiveUserID != "" {
		tmc.WithUserID(effectiveUserID)
	}

	if o.ResourceID != "" {
		tmc.WithResourceID(o.ResourceID)
	}

	teams, err := c.Teams(tmc)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(teams))
	for i, team := range teams {
		ids[i] = team.ID
	}

	return ids, nil
}

// Create creates a new team
//
// Notifications are sent out if the owner of the team is not admin
func (c *Client) Create(id, ownerID string, dtoCreateTeam *body.TeamCreate) (*model.Team, error) {
	params := &model.TeamCreateParams{}
	params.FromDTO(dtoCreateTeam, ownerID,
		func(resourceID string) *model.TeamResource { return c.getResourceIfAccessible(resourceID) },
		func(memberDTO *body.TeamMemberCreate) *model.TeamMember {
			return c.createMemberIfAccessible(nil, memberDTO.ID)
		},
	)

	team, err := team_repo.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, team_repo.NameTakenErr) {
			return nil, sErrors.TeamNameTakenErr
		}

		return nil, err
	}

	// Send invitations to every member that received an invitation code
	for _, member := range params.MemberMap {
		if member.InvitationCode != "" {
			err = c.createInvitationNotification(member.ID, team.ID, team.Name, member.InvitationCode)
			if err != nil {
				return nil, err
			}
		}
	}

	return team, nil
}

// Update updates a team
func (c *Client) Update(id string, dtoUpdateTeam *body.TeamUpdate) (*model.Team, error) {
	team, err := team_repo.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	if c.V2.Auth() != nil && team.OwnerID != c.V2.Auth().User.ID && !c.V2.Auth().User.IsAdmin && !team.HasMember(c.V2.Auth().User.ID) {
		return nil, nil
	}

	params := &model.TeamUpdateParams{}
	params.FromDTO(dtoUpdateTeam, team.GetMember(team.OwnerID),
		func(resourceID string) *model.TeamResource { return c.getResourceIfAccessible(resourceID) },
		func(memberDTO *body.TeamMemberUpdate) *model.TeamMember {
			return c.createMemberIfAccessible(team, memberDTO.ID)
		},
	)

	// If new user, set timestamp
	if params.MemberMap != nil {
		for _, member := range *params.MemberMap {
			// Don't invite users that have already joined
			if existing := team.GetMember(member.ID); existing != nil && existing.MemberStatus == model.TeamMemberStatusJoined {
				continue
			}

			// Send notification to new users
			if member.InvitationCode != "" {
				err = c.createInvitationNotification(member.ID, team.ID, team.Name, member.InvitationCode)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// If new model, set timestamp
	if params.ResourceMap != nil {
		for _, resource := range *params.ResourceMap {
			if existing := team.GetResource(resource.ID); existing != nil {
				resource.AddedAt = existing.AddedAt
			} else {
				resource.AddedAt = time.Now()
			}

			(*params.ResourceMap)[resource.ID] = resource
		}
	}

	tmc := team_repo.New()

	err = tmc.UpdateWithParams(id, params)
	if err != nil {
		return nil, err
	}

	return c.RefreshTeam(id, tmc)
}

// Delete deletes a team
func (c *Client) Delete(id string) error {
	tmc := team_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		tmc.WithOwnerID(c.V2.Auth().User.ID)
	}

	exists, err := tmc.ExistsByID(id)
	if err != nil {
		return err
	}

	if !exists {
		return sErrors.TeamNotFoundErr
	}

	err = notification_repo.New().FilterContent("id", id).Delete()
	if err != nil {
		return err
	}

	return tmc.DeleteByID(id)
}

// CleanResource cleans a resource from all teams
func (c *Client) CleanResource(resourceID string) error {
	tmc := team_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		tmc.WithOwnerID(c.V2.Auth().User.ID)
	}

	tmc.WithResourceID(resourceID)

	teams, err := c.Teams(tmc)
	if err != nil {
		return err
	}

	for _, team := range teams {
		delete(team.GetResourceMap(), resourceID)
		err = tmc.UpdateWithBsonByID(team.ID, bson.D{{"$set", bson.D{{"resourceMap", team.ResourceMap}}}})
		if err != nil {
			return err
		}
	}

	return nil
}

// Join joins a team
//
// It uses service.AuthInfo to get the user ID
func (c *Client) Join(id string, dtoTeamJoin *body.TeamJoin) (*model.Team, error) {
	if !c.V2.HasAuth() {
		return nil, nil
	}

	params := &model.TeamJoinParams{}
	params.FromDTO(dtoTeamJoin)

	tmc := team_repo.New()
	team, err := tmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	if team.GetMemberMap()[c.V2.Auth().User.ID].MemberStatus != model.TeamMemberStatusInvited {
		return team, sErrors.NotInvitedErr
	}

	if team.GetMemberMap()[c.V2.Auth().User.ID].InvitationCode != params.InvitationCode {
		return nil, sErrors.BadInviteCodeErr
	}

	updatedMember := team.GetMemberMap()[c.V2.Auth().User.ID]
	updatedMember.MemberStatus = model.TeamMemberStatusJoined
	updatedMember.JoinedAt = time.Now()

	err = tmc.UpdateMember(id, c.V2.Auth().User.ID, &updatedMember)
	if err != nil {
		return nil, err
	}

	nmc := notification_repo.New().WithUserID(c.V2.Auth().User.ID).FilterContent("id", id).WithType(model.NotificationTeamInvite)
	err = nmc.MarkReadAndCompleted()
	if err != nil {
		return nil, err
	}

	return c.RefreshTeam(id, tmc)
}

func (c *Client) CheckResourceAccess(userID, resourceID string) (bool, error) {
	return team_repo.New().WithUserID(userID).WithResourceID(resourceID).ExistsAny()
}

// getTeamIfAccessible is a helper function to get a team if the user is accessible to the user in the current context
func (c *Client) getResourceIfAccessible(resourceID string) *model.TeamResource {
	// Try to fetch deployment
	dmc := deployment_repo.New()
	vmc := vm_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		dmc.WithOwner(c.V2.Auth().User.ID)
		vmc.WithOwner(c.V2.Auth().User.ID)
	}

	isOwner, err := dmc.ExistsByID(resourceID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to fetch deployment when checking user access when creating team: %w", err))
		return nil
	}

	if isOwner {
		return &model.TeamResource{
			ID:      resourceID,
			Type:    model.ResourceTypeDeployment,
			AddedAt: time.Now(),
		}
	}

	// Try to fetch vm
	isOwner, err = vm_repo.New().ExistsByID(resourceID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to fetch vm when checking user access when creating team: %w", err))
		return nil
	}

	if isOwner {
		return &model.TeamResource{
			ID:      resourceID,
			Type:    model.ResourceTypeVM,
			AddedAt: time.Now(),
		}
	}

	return nil
}

// createMemberIfAccessible is a helper function to create a member for a team if the user is accessible
// to the user in the current context
func (c *Client) createMemberIfAccessible(current *model.Team, memberID string) *model.TeamMember {
	if current != nil {
		if existing := current.GetMember(memberID); existing != nil {
			existing.TeamRole = model.TeamMemberRoleAdmin
			return existing
		}
	}

	member := &model.TeamMember{
		ID:       memberID,
		TeamRole: model.TeamMemberRoleAdmin,
		AddedAt:  time.Now(),
	}

	if !c.V2.HasAuth() || c.V2.Auth().User.IsAdmin {
		member.MemberStatus = model.TeamMemberStatusJoined
		member.JoinedAt = time.Now()
	} else {
		member.MemberStatus = model.TeamMemberStatusInvited
		member.InvitationCode = createInvitationCode()
	}

	return member
}

// createInvitationNotification is a helper function to create a notification for a team invitation
func (c *Client) createInvitationNotification(userID, teamID, teamName, invitationCode string) error {
	_, err := c.V2.Notifications().Create(uuid.NewString(), userID, &model.NotificationCreateParams{
		Type: model.NotificationTeamInvite,
		Content: map[string]interface{}{
			"id":   teamID,
			"name": teamName,
			"code": invitationCode,
		},
	})

	return err
}

// createInvitationCode is a helper function to create a random invitation code
func createInvitationCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
