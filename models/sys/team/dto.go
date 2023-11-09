package team

import (
	"go-deploy/models/dto/body"
)

func (t *Team) ToDTO(getMember func(*Member) *body.TeamMember, getResourceName func(*Resource) *string) body.TeamRead {
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
		if memberDTO := getMember(&member); memberDTO != nil {
			members = append(members, *memberDTO)
		}
	}

	return body.TeamRead{
		ID:          t.ID,
		Name:        t.Name,
		OwnerID:     t.OwnerID,
		Description: t.Description,
		Resources:   resources,
		Members:     members,
	}
}

func (params *CreateParams) FromDTO(teamCreateDTO *body.TeamCreate, getResourceFunc func(string) *Resource) {
	params.Name = teamCreateDTO.Name
	params.MemberMap = make(map[string]Member)

	for _, resourceDTO := range teamCreateDTO.Resources {
		if resource := getResourceFunc(resourceDTO); resource != nil {
			params.ResourceMap[resource.ID] = *resource
		}
	}

	for _, member := range teamCreateDTO.Members {
		params.MemberMap[member.ID] = Member{
			ID:       member.ID,
			TeamRole: member.TeamRole,
		}
	}
}

func (params *JoinParams) FromDTO(teamJoinDTO *body.TeamJoin) {
	params.InvitationCode = teamJoinDTO.InvitationCode
}

func (params *UpdateParams) FromDTO(teamUpdateDTO *body.TeamUpdate, getResourceFunc func(string) *Resource) {
	params.Name = teamUpdateDTO.Name
	params.Description = teamUpdateDTO.Description

	if teamUpdateDTO.Resources != nil {
		resourceMap := make(map[string]Resource)
		for _, resourceDTO := range *teamUpdateDTO.Resources {
			if resource := getResourceFunc(resourceDTO); resource != nil {
				resourceMap[resource.ID] = *resource
			}
		}
		params.ResourceMap = &resourceMap
	}

	if teamUpdateDTO.Members != nil {
		memberMap := make(map[string]Member)
		for _, member := range *teamUpdateDTO.Members {
			memberMap[member.ID] = Member{
				ID: member.ID,
				// temporary until we have a use case for this
				TeamRole: MemberRoleAdmin,
			}
		}
		params.MemberMap = &memberMap
	}
}
