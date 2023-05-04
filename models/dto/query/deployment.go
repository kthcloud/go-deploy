package query

type DeploymentList struct {
	WantAll bool `form:"all" binding:"omitempty,boolean"`
}
