package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
)

func (client *Client) AssertPortForwardingRulesTags(resourceID string, tags []models.Tag) error {
	return client.AssertTags(resourceID, "PortForwardingRule", tags)
}

func (client *Client) AssertVirtualMachineTags(resourceID string, tags []models.Tag) error {
	return client.AssertTags(resourceID, "UserVm", tags)
}

func (client *Client) AssertPublicIPAddressTags(resourceID string, tags []models.Tag) error {
	return client.AssertTags(resourceID, "PublicIpAddress", tags)
}

func (client *Client) AssertTags(resourceID, resourceType string, tags []models.Tag) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create tags for resource %s. details: %w", resourceID, err)
	}

	listTagsParams := client.CsClient.Resourcetags.NewListTagsParams()
	listTagsParams.SetResourceid(resourceID)
	listTagsParams.SetResourcetype(resourceType)
	listTagsParams.SetListall(true)
	listTagsParams.SetProjectid(client.ProjectID)

	currentTags, err := client.CsClient.Resourcetags.ListTags(listTagsParams)
	if err != nil {
		return makeError(err)
	}

	// check if any of the tags are not already present
	var notFound []models.Tag
	for _, tag := range tags {
		found := false
		for _, currentTag := range currentTags.Tags {
			if currentTag.Key == tag.Key {
				found = true
				break
			}
		}

		if !found {
			notFound = append(notFound, tag)
		}
	}

	if len(notFound) == 0 {
		return nil
	}

	csTags := map[string]string{}
	for _, tag := range notFound {
		csTags[tag.Key] = tag.Value
	}

	createTagParams := client.CsClient.Resourcetags.NewCreateTagsParams(
		[]string{resourceID},
		resourceType,
		csTags,
	)

	_, err = client.CsClient.Resourcetags.CreateTags(createTagParams)
	if err != nil {
		return makeError(err)
	}

	return nil
}
