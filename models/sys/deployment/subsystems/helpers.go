package subsystems

func (k8s *K8s) Created() bool {
	return k8s.Namespace.Created() &&
		k8s.Deployment.Created() &&
		k8s.Service.Created() &&
		k8s.Ingress.Created()
}

func (harbor *Harbor) Created() bool {
	return harbor.Project.Created() &&
		harbor.Repository.Created() &&
		harbor.Robot.Created() &&
		harbor.Webhook.Created()
}

func (gitHub *GitHub) Created() bool {
	return gitHub.Webhook.Created()
}
