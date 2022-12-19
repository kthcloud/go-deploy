package models

type PlaceHolder struct {
	ProjectName    string `json:"projectName" bson:"projectName"`
	RepositoryName string `json:"repositoryName" bson:"repositoryName"`
}
