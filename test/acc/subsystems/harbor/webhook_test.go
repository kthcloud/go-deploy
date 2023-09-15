package harbor

import "testing"

func TestCreateWebhook(t *testing.T) {
	project := withHarborProject(t)
	webhook := withHarborWebhook(t, project)
	cleanUpHarborWebhook(t, webhook.ID, project.ID)
	cleanUpHarborProject(t, project.ID)
}
