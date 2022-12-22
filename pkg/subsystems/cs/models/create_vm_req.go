package models

type VMPublic struct {
	ID                string
	Name              string
	ServiceOfferingID string
	TemplateID        string
	NetworkID         string
	ZoneID            string
	ProjectID         string
	ExtraConfig       string
}
