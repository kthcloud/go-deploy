package model

import (
	"time"
)

var (
	// FetchGravatarInterval is the interval at which we fetch the gravatar
	// image for a user. This is done to prevent fetching the image on every
	// request.
	FetchGravatarInterval = 10 * time.Minute
)

const (
	TestAdminUserID   = "955f0f87-37fd-4792-90eb-9bf6989e698a"
	TestPowerUserID   = "955f0f87-37fd-4792-90eb-9bf6989e698b"
	TestDefaultUserID = "955f0f87-37fd-4792-90eb-9bf6989e698c"

	TestAdminUserApiKey   = "test-api-key-admin"
	TestPowerUserApiKey   = "test-api-key-power"
	TestDefaultUserApiKey = "test-api-key-default"
)

type User struct {
	ID        string   `bson:"id"`
	Username  string   `bson:"username"`
	FirstName string   `bson:"firstName"`
	LastName  string   `bson:"lastName"`
	Email     string   `bson:"email"`
	Gravatar  Gravatar `bson:"gravatar"`

	IsAdmin       bool          `bson:"isAdmin"`
	EffectiveRole EffectiveRole `bson:"effectiveRole"`

	PublicKeys []PublicKey `bson:"publicKeys,omitempty"`
	ApiKeys    []ApiKey    `bson:"apiKeys,omitempty"`
	UserData   []UserData  `bson:"userData,omitempty"`

	LastAuthenticatedAt time.Time `bson:"lastAuthenticatedAt"`
}

type AuthParams struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"isAdmin"`
	Roles     []Role `json:"roles"`
}

type PublicKey struct {
	Name string `bson:"name"`
	Key  string `bson:"key"`
}

type UserData struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
}

type UserUsage struct {
	CpuCores  float64 `bson:"cpuCores"`
	RAM       float64 `bson:"ram"`
	DiskSize  int     `bson:"diskSize"`
	Snapshots int     `bson:"snapshots"`
}

type EffectiveRole struct {
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

type Gravatar struct {
	// URL is the URL to the gravatar image.
	// If the URL is nil, the gravatar image has not been fetched yet, or the
	// user does not have a gravatar image.
	URL       *string   `bson:"url"`
	FetchedAt time.Time `bson:"fetchedAt"`
}

func CreateEmptyGravatar() Gravatar {
	return Gravatar{
		URL:       nil,
		FetchedAt: time.Time{},
	}
}
