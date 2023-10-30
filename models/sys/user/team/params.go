package team

type CreateParams struct {
	Name        string
	Description string
	ResourceMap map[string]Resource
	MemberMap   map[string]Member
}

type JoinParams struct {
	InvitationCode string
}

type UpdateParams struct {
	Name        *string
	Description *string
	MemberMap   *map[string]Member
	ResourceMap *map[string]Resource
}
