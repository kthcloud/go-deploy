package team

type CreateParams struct {
	Name        string              `bson:"name"`
	Description string              `bson:"description"`
	ResourceMap map[string]Resource `bson:"resourceMap"`
	MemberMap   map[string]Member   `bson:"memberMap"`
}

type UpdateParams struct {
	Name        *string              `bson:"name"`
	Description *string              `bson:"description"`
	MemberMap   *map[string]Member   `bson:"memberMap"`
	ResourceMap *map[string]Resource `bson:"resourceMap"`
}
