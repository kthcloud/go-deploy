package query

type ZoneList struct {
	Type *string `form:"type" binding:"omitempty,oneof=deployment vm"`
}
