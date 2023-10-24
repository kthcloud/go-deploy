package team

import (
	"go-deploy/models/dto/body"
)

func (t *Team) ToDTO(users map[string]body.UserRead) body.TeamRead {
	members := make([]body.TeamMember, 0)

	userIDs := make([]string, len(t.MemberMap))
	i := 0
	for userID := range t.MemberMap {
		userIDs[i] = userID
		i++
	}

	for _, member := range t.GetMemberMap() {
		if user, ok := users[member.ID]; ok {
			members = append(members, body.TeamMember{
				UserRead: user,
				TeamRole: member.TeamRole,
				JoinedAt: member.JoinedAt,
			})
			continue
		}
	}

	return body.TeamRead{
		ID:      t.ID,
		Name:    t.Name,
		Members: members,
	}
}

func (params *CreateParams) FromDTO(teamCreateDTO *body.TeamCreate) {
	params.Name = teamCreateDTO.Name
	params.MemberMap = make(map[string]Member)

	for _, user := range teamCreateDTO.Members {
		params.MemberMap[user.ID] = Member{
			ID:       user.ID,
			TeamRole: user.TeamRole,
		}
	}
}

func (params *UpdateParams) FromDTO(teamUpdateDTO *body.TeamUpdate) {
	params.Name = teamUpdateDTO.Name

	if teamUpdateDTO.Members != nil {
		memberMap := make(map[string]Member)
		for _, member := range *teamUpdateDTO.Members {
			memberMap[member.ID] = Member{
				ID:       member.ID,
				TeamRole: member.TeamRole,
			}
		}
		params.MemberMap = &memberMap
	}
}
