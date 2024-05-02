package uri

type UserGet struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}

type UserUpdate struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}

type ApiKeyCreate struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}
