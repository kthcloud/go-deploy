package team_repo

import (
	"errors"
	"fmt"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var (
	NameTakenErr = fmt.Errorf("team name taken")
)

// Create creates a new team.
func (client *Client) Create(id, ownerID string, params *model.TeamCreateParams) (*model.Team, error) {
	team := &model.Team{
		ID:          id,
		Name:        params.Name,
		Description: params.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		ResourceMap: params.ResourceMap,
		MemberMap:   params.MemberMap,
	}

	err := client.CreateIfUnique(id, team, bson.D{{"name", params.Name}})
	if err != nil {
		if errors.Is(err, db.UniqueConstraintErr) {
			return nil, NameTakenErr
		}
		return nil, err
	}

	fetched, err := client.GetByName(params.Name)
	if err != nil {
		return nil, err
	}

	return fetched, nil
}

func (client *Client) ListMemberIDs(ids ...string) ([]string, error) {
	teams, err := client.ListWithFilterAndProjection(bson.D{{"id", bson.D{{"$in", ids}}}}, bson.D{{"memberMap", 1}})
	if err != nil {
		return nil, err
	}

	memberIDs := make(map[string]bool, 0)
	for _, team := range teams {
		for memberID := range team.MemberMap {
			if _, ok := memberIDs[memberID]; !ok {
				memberIDs[memberID] = true
			}
		}
	}

	result := make([]string, 0, len(memberIDs))
	for memberID := range memberIDs {
		result = append(result, memberID)
	}

	return result, nil
}

func (client *Client) UpdateWithParams(id string, params *model.TeamUpdateParams) error {
	updateData := bson.D{
		{"updatedAt", time.Now()},
	}

	db.AddIfNotNil(&updateData, "name", params.Name)
	db.AddIfNotNil(&updateData, "description", params.Description)
	db.AddIfNotNil(&updateData, "resourceMap", params.ResourceMap)
	db.AddIfNotNil(&updateData, "memberMap", params.MemberMap)

	if len(updateData) == 0 {
		return nil
	}

	return client.SetWithBsonByID(id, updateData)
}

func (client *Client) UpdateMember(id string, memberID string, member *model.TeamMember) error {
	updateData := bson.D{
		{"updatedAt", time.Now()},
	}

	db.AddIfNotNil(&updateData, "memberMap."+memberID, member)

	if len(updateData) == 0 {
		return nil
	}

	return client.SetWithBsonByID(id, updateData)
}
