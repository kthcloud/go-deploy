package query

type DeploymentList struct {
	WantAll bool `query:"all" binding:"omitempty,boolean"`
}
