package subsystems

func (gitHub *GitHub) Created() bool {
	return gitHub.Webhook.Created()
}
