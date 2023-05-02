package uri

type UserGet struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}
