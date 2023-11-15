package team

import (
	"go-deploy/models/dto/body"
	"go-deploy/utils"
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
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   utils.NonZeroOrNil(t.UpdatedAt),
	}
}

func (params *CreateParams) FromDTO(teamCreateDTO *body.TeamCreate, getResourceFunc func(string) *Resource, getMemberFunc func(*body.TeamMemberCreate) *Member) {
	params.Name = teamCreateDTO.Name
	params.MemberMap = make(map[string]Member)

	for _, resourceDTO := range teamCreateDTO.Resources {
		if resource := getResourceFunc(resourceDTO); resource != nil {
			params.ResourceMap[resource.ID] = *resource
		}
	}

	for _, memberDTO := range teamCreateDTO.Members {
		params.MemberMap[memberDTO.ID] = *getMemberFunc(&memberDTO)
	}
}

func (params *JoinParams) FromDTO(teamJoinDTO *body.TeamJoin) {
	params.InvitationCode = teamJoinDTO.InvitationCode
}

func (params *UpdateParams) FromDTO(teamUpdateDTO *body.TeamUpdate, getResourceFunc func(string) *Resource, getMemberFunc func(*body.TeamMemberUpdate) *Member) {
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
		for _, memberDTO := range *teamUpdateDTO.Members {
			memberMap[memberDTO.ID] = *getMemberFunc(&memberDTO)
		}
		params.MemberMap = &memberMap
	}
}
