package body

type DeploymentCreate struct {
	Name string `json:"name" binding:"required,rfc1035,min=3,max=30"`
}
