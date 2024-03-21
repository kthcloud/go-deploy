package model

import (
	"go-deploy/dto/v1/body"
	"go-deploy/utils"
	"sort"
	"time"
)

// ToDTO converts a Team to a body.TeamRead DTO.
func (t *Team) ToDTO(getMember func(*TeamMember) *body.TeamMember, getResourceName func(*TeamResource) *string) body.TeamRead {
	resources := make([]body.TeamResource, 0)
	members := make([]body.TeamMember, 0)

	for _, resource := range t.GetResourceMap() {
		if resourceName := getResourceName(&resource); resourceName != nil {
			resources = append(resources, body.TeamResource{
				ID:   resource.ID,
				Name: *resourceName,
				Type: resource.Type,
			})
		}
	}

	for _, member := range t.GetMemberMap() {
		lMember := member
		if memberDTO := getMember(&lMember); memberDTO != nil {
			l := *memberDTO
			members = append(members, l)
		}
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	sort.Slice(members, func(i, j int) bool {
		return members[i].Username < members[j].Username
	})

	return body.TeamRead{
		ID:          t.ID,
		Name:        t.Name,
		OwnerID:     t.OwnerID,
		Description: t.Description,
		Resources:   resources,
		Members:     members,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   utils.NonZeroOrNil(t.UpdatedAt),
	}
}

// FromDTO converts a body.TeamCreate DTO to a NotificationCreateParams.
func (params *TeamCreateParams) FromDTO(teamCreateDTO *body.TeamCreate, ownerID string, getResourceFunc func(string) *TeamResource, getMemberFunc func(*body.TeamMemberCreate) *TeamMember) {
	params.Name = teamCreateDTO.Name
	params.MemberMap = make(map[string]TeamMember)
	params.Description = teamCreateDTO.Description
	params.ResourceMap = make(map[string]TeamResource)

	for _, resourceDTO := range teamCreateDTO.Resources {
		if resource := getResourceFunc(resourceDTO); resource != nil {
			params.ResourceMap[resource.ID] = *resource
		}
	}

	now := time.Now()

	params.MemberMap[ownerID] = TeamMember{
		ID:           ownerID,
		TeamRole:     TeamMemberRoleAdmin,
		AddedAt:      now,
		JoinedAt:     now,
		MemberStatus: TeamMemberStatusJoined,
	}

	for _, memberDTO := range teamCreateDTO.Members {
		params.MemberMap[memberDTO.ID] = *getMemberFunc(&memberDTO)
	}
}

// FromDTO converts a body.TeamJoin DTO to a JoinParams.
func (params *TeamJoinParams) FromDTO(teamJoinDTO *body.TeamJoin) {
	params.InvitationCode = teamJoinDTO.InvitationCode
}

// FromDTO converts a body.TeamUpdate DTO to an UpdateParams.
func (params *TeamUpdateParams) FromDTO(teamUpdateDTO *body.TeamUpdate, owner *TeamMember, getResourceFunc func(string) *TeamResource, getMemberFunc func(*body.TeamMemberUpdate) *TeamMember) {
	params.Name = teamUpdateDTO.Name
	params.Description = teamUpdateDTO.Description

	if teamUpdateDTO.Resources != nil {
		resourceMap := make(map[string]TeamResource)
		for _, resourceDTO := range *teamUpdateDTO.Resources {
			if resource := getResourceFunc(resourceDTO); resource != nil {
				resourceMap[resource.ID] = *resource
			}
		}
		params.ResourceMap = &resourceMap
	}

	if teamUpdateDTO.Members != nil {
		memberMap := make(map[string]TeamMember)

		memberMap[owner.ID] = *owner

		for _, memberDTO := range *teamUpdateDTO.Members {
			memberMap[memberDTO.ID] = *getMemberFunc(&memberDTO)
		}
		params.MemberMap = &memberMap
	}
}
