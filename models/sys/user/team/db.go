package team

import (
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var NameTakenErr = fmt.Errorf("team name taken")

func (client *Client) Create(id, ownerID string, params *CreateParams) (*Team, error) {
	team := &Team{
		ID:          id,
		Name:        params.Name,
		Description: &params.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		ResourceMap: params.ResourceMap,
		MemberMap:   params.MemberMap,
	}

	err := client.CreateIfUnique(id, team, bson.D{{"name", params.Name}})
	if err != nil {
		if err == models.UniqueConstraintErr {
			return nil, NameTakenErr
		} else {
			return nil, err
		}
	}

	fetched, err := client.GetByName(params.Name)
	if err != nil {
		return nil, err
	}

	return fetched, nil
}

func (client *Client) GetByIdList(ids []string) ([]Team, error) {
	if ids == nil || len(ids) == 0 {
		return make([]Team, 0), nil
	}

	filter := bson.D{{"id", bson.D{{"$in", ids}}}}

	return client.GetAllWithFilter(filter)
}

func (client *Client) GetByUserID(userID string) ([]Team, error) {
	filter := bson.D{{"userMap." + userID, bson.D{{"$exists", true}}}}
	return client.GetAllWithFilter(filter)
}

func (client *Client) UpdateWithParamsByID(id string, params *UpdateParams) error {
	updateData := bson.D{}

	if params.MemberMap != nil {
		for _, member := range *params.MemberMap {
			if member.JoinedAt.IsZero() {
				member.JoinedAt = time.Now()
			} else {
				member.JoinedAt = time.Time{} // reset to zero value so it doesn't get updated
			}
		}
	}

	models.AddIfNotNil(&updateData, "name", params.Name)
	models.AddIfNotNil(&updateData, "description", params.Description)
	models.AddIfNotNil(&updateData, "resourceMap", params.ResourceMap)
	models.AddIfNotNil(&updateData, "memberMap", params.MemberMap)

	if len(updateData) == 0 {
		return nil
	}

	err := client.SetWithBsonByID(id, updateData)
	return err
}
