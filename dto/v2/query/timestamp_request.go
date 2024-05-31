package query

type TimestampRequest struct {
	N int `form:"n" binding:"omitempty,numeric,min=1,max=1000"`
}
