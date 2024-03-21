package uri

type UserDataGet struct {
	ID string `uri:"id" binding:"required,rfc1035"`
}

type UserDataUpdate struct {
	ID string `uri:"id" binding:"required,rfc1035"`
}

type UserDataDelete struct {
	ID string `uri:"id" binding:"required,rfc1035"`
}
