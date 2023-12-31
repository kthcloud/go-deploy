package cs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"testing"
)

func TestCreateServiceOffering(t *testing.T) {
	t.Parallel()

	withCsServiceOfferingSmall(t)
	withCsServiceOfferingBig(t)
}

func TestUpdateServiceOffering(t *testing.T) {
	t.Parallel()

	client := withClient(t)
	so := withCsServiceOfferingSmall(t)

	so.Name = so.Name + "-updated"
	so.Description = so.Description + "-updated"

	soUpdated, err := client.UpdateServiceOffering(so)
	test.NoError(t, err, "failed to update service offering")

	assert.Equal(t, so.Name, soUpdated.Name, "service offering name is not updated")
	assert.Equal(t, so.Description, soUpdated.Description, "service offering description is not updated")
}
