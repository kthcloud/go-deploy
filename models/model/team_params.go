package model

type TeamCreateParams struct {
	Name        string
	Description string
	ResourceMap map[string]TeamResource
	MemberMap   map[string]TeamMember
}

type TeamJoinParams struct {
	InvitationCode string
}

type TeamUpdateParams struct {
	Name        *string
	Description *string
	MemberMap   *map[string]TeamMember
	ResourceMap *map[string]TeamResource
}
