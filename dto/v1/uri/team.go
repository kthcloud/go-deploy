package uri

type TeamGet struct {
	TeamID string `uri:"teamId" binding:"omitempty,uuid4"`
}

type TeamUpdate struct {
	TeamID string `uri:"teamId" binding:"omitempty,uuid4"`
}
