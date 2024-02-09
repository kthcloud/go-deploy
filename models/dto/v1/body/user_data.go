package body

type UserDataRead struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type UserDataCreate struct {
	ID   string `json:"id" binding:"required,min=1,max=100"`
	Data string `json:"data" binding:"required,min=1,max=1000"`
}

type UserDataUpdate struct {
	Data string `json:"data" binding:"required,min=1,max=1000"`
}
