package constants

const (
	AppName                = "main"
	AppNameImagePullSecret = "image-pull-secret"
	AppNameCustomDomain    = "custom-domain"
	VmProxyAppName         = "vm-proxy"
	SmNamePrefix           = "system"
	SmAppName              = "storage-manager"
	SmAppNameAuth          = "storage-manager-auth"
	WildcardCertSecretName = "wildcard-cert"
)

func WithImagePullSecretSuffix(appName string) string {
	return appName + "-" + AppNameImagePullSecret
}

func WithCustomDomainSuffix(appName string) string {
	return appName + "-" + AppNameCustomDomain
}
