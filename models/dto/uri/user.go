package uri

type UserGet struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}

type UserUpdate struct {
	UserID string `uri:"userId" binding:"omitempty,uuid4"`
}

type TeamGet struct {
	TeamID string `uri:"teamId" binding:"omitempty,uuid4"`
}

type TeamUpdate struct {
	TeamID string `uri:"teamId" binding:"omitempty,uuid4"`
}
