package body

import "time"

type TeamMember struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`

	TeamRole     string `json:"teamRole"`
	MemberStatus string `json:"memberStatus"`

	JoinedAt *time.Time `json:"joinedAt,omitempty"`
	AddedAt  *time.Time `json:"addedAt,omitempty"`
}

type TeamResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TeamMemberCreate struct {
	ID       string `json:"id" binding:"required,uuid4"`
	TeamRole string `json:"teamRole" binding:"omitempty"` // default to MemberRoleAdmin right now
}

type TeamMemberUpdate struct {
	ID       string `json:"id" binding:"required,uuid4"`
	TeamRole string `json:"teamRole" binding:"omitempty"` // default to MemberRoleAdmin right now
}

type TeamCreate struct {
	Name        string             `json:"name" binding:"required,min=1,max=100"`
	Description string             `json:"description" binding:"omitempty,max=1000"`
	Resources   []string           `json:"resources" binding:"omitempty,team_resource_list,min=0,max=10,dive,uuid4"`
	Members     []TeamMemberCreate `json:"members" binding:"omitempty,team_member_list,min=0,max=10,dive"`
}

type TeamJoin struct {
	InvitationCode string `json:"invitationCode" binding:"required,min=1,max=1000"`
}

type TeamUpdate struct {
	Name        *string             `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string             `json:"description,omitempty" binding:"omitempty,max=1000"`
	Resources   *[]string           `json:"resources,omitempty" binding:"omitempty,team_resource_list,min=0,max=10,dive,uuid4"`
	Members     *[]TeamMemberUpdate `json:"members,omitempty" binding:"omitempty,team_member_list,min=0,max=10,dive"`
}

type TeamRead struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	OwnerID     string         `json:"ownerId"`
	Description *string        `json:"description,omitempty"`
	Resources   []TeamResource `json:"resources"`
	Members     []TeamMember   `json:"members"`
}
