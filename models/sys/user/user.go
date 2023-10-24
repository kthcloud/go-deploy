package user

import (
	"fmt"
	teamModel "go-deploy/models/sys/user/team"
)

type PublicKey struct {
	Name string `bson:"name"`
	Key  string `bson:"key"`
}

type Usage struct {
	Deployments int `bson:"deployments"`
	CpuCores    int `bson:"cpuCores"`
	RAM         int `bson:"ram"`
	DiskSize    int `bson:"diskSize"`
	Snapshots   int `bson:"snapshots"`
}

type EffectiveRole struct {
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

type User struct {
	ID            string        `bson:"id"`
	Username      string        `bson:"username"`
	Email         string        `bson:"email"`
	IsAdmin       bool          `bson:"isAdmin"`
	EffectiveRole EffectiveRole `bson:"effectiveRole"`
	PublicKeys    []PublicKey   `bson:"publicKeys"`
	Onboarded     bool          `bson:"onboarded"`
}

func (u *User) GetTeamMap() (map[string]teamModel.Team, error) {
	client := teamModel.New()

	teams, err := client.GetByUserID(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user %s. details: %w", u.Username, err)
	}

	teamsMap := make(map[string]teamModel.Team)
	for _, team := range teams {
		teamsMap[team.ID] = team
	}

	return teamsMap, nil
}
