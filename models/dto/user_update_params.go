package dto

type UserUpdateParams struct {
	UserID *string `uri:"userId" binding:"omitempty,uuid4"`
}
