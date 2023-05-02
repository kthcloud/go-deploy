package uri

type UserUpdate struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}
