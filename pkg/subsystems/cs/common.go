package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/errors"
	"go-deploy/pkg/subsystems/cs/models"
)

const (
	// DetailsCpuCores is the key for the number of CPU cores.
	DetailsCpuCores = "cpuNumber"
	// DetailsCpuSpeed is the key for the CPU speed.
	DetailsCpuSpeed = "cpuSpeed"
	// DetailsRAM is the key for the memory.
	DetailsRAM = "memory"
)

// AssertPortForwardingRulesTags asserts that the tags are present on the port forwarding rule.
func (client *Client) AssertPortForwardingRulesTags(resourceID string, tags []models.Tag) error {
	return client.AssertTags(resourceID, "PortForwardingRule", tags)
}

// AssertVmTags asserts that the tags are present on the virtual machine.
func (client *Client) AssertVmTags(resourceID string, tags []models.Tag) error {
	return client.AssertTags(resourceID, "UserVm", tags)
}

// AssertPublicIpAddressTags asserts that the tags are present on the public IP address.
func (client *Client) AssertPublicIpAddressTags(resourceID string, tags []models.Tag) error {
	return client.AssertTags(resourceID, "PublicIpAddress", tags)
}

// AssertTags asserts that the tags are present on the resource.
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

// getCustomServiceOfferingID returns the ID of the custom service offering.
// The `custom` service offering is a service offering in CloudStack that is able to take custom
// custom CPU cores, CPU speed and memory.
func (client *Client) getCustomServiceOfferingID() (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get custom service offering id: %w", err)
	}

	listServiceOfferingsParams := client.CsClient.ServiceOffering.NewListServiceOfferingsParams()
	listServiceOfferingsParams.SetName("custom")
	listServiceOfferingsParams.SetListall(true)

	serviceOfferings, err := client.CsClient.ServiceOffering.ListServiceOfferings(listServiceOfferingsParams)
	if err != nil {
		return "", makeError(err)
	}

	if len(serviceOfferings.ServiceOfferings) == 0 {
		return "", errors.NotFoundErr
	}

	return serviceOfferings.ServiceOfferings[0].Id, nil
}
