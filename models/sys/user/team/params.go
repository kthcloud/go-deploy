package team

type CreateParams struct {
	Name      string            `bson:"name"`
	MemberMap map[string]Member `bson:"memberMap"`
}

type UpdateParams struct {
	Name      *string            `bson:"name"`
	MemberMap *map[string]Member `bson:"memberMap"`
}
