package uri

type ZoneGet struct {
	Name string `uri:"name" binding:"required,rfc1035,min=1,max=30"`
	Type string `uri:"type" binding:"required,oneof=deployment vm"`
}
