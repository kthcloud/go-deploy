package team

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var NameTaken = fmt.Errorf("team name taken")

func (client *Client) Create(id string, params *CreateParams) (*Team, error) {
	team := Team{
		ID:        id,
		Name:      params.Name,
		MemberMap: params.MemberMap,
	}

	filter := bson.D{{"name", params.Name}, {"deletedAt", bson.D{{"$in", []interface{}{time.Time{}, nil}}}}}
	result, err := client.Collection.UpdateOne(context.TODO(), filter, bson.D{
		{"$setOnInsert", team},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create team. details: %w", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetchedDeployment, err := client.GetByName(params.Name)
			if err != nil {
				return nil, err
			}

			if fetchedDeployment == nil {
				log.Println(fmt.Errorf("failed to fetch team %s after creation. assuming it was deleted", params.Name))
				return nil, nil
			}

			if fetchedDeployment.ID == id {
				return fetchedDeployment, nil
			}
		}

		return nil, NameTaken
	}

	fetchedDeployment, err := client.GetByName(params.Name)
	if err != nil {
		return nil, err
	}

	return fetchedDeployment, nil
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
			member.JoinedAt = time.Time{} // reset to zero value so it doesn't get updated
		}
	}

	models.AddIfNotNil(&updateData, "name", params.Name)
	models.AddIfNotNil(&updateData, "userMap", params.MemberMap)

	if len(updateData) == 0 {
		return nil
	}

	return client.UpdateWithBsonByID(id, updateData)
}

func (client *Client) HasUser(teamID string) (bool, error) {
	filter := bson.D{{"users." + teamID, bson.D{{"$exists", true}}}}

	count, err := client.Collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, fmt.Errorf("failed to check if user %s exists. details: %w", teamID, err)
	}

	return count > 0, nil
}
