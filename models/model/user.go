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

type User struct {
	ID        string   `bson:"id"`
	Username  string   `bson:"username"`
	FirstName string   `bson:"firstName"`
	LastName  string   `bson:"lastName"`
	Email     string   `bson:"email"`
	Gravatar  Gravatar `bson:"gravatar"`

	IsAdmin       bool          `bson:"isAdmin"`
	EffectiveRole EffectiveRole `bson:"effectiveRole"`

	PublicKeys []PublicKey `bson:"publicKeys"`
	ApiKeys    []ApiKey    `bson:"apiKeys"`
	UserData   []UserData  `bson:"userData"`

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
	DiskSize  float64 `bson:"diskSize"`
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
