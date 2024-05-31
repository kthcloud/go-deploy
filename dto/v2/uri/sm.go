package uri

type SmGet struct {
	SmID string `uri:"storageManagerId" binding:"required,uuid4"`
}

type SmDelete struct {
	SmID string `uri:"storageManagerId" binding:"required,uuid4"`
}
