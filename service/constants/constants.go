package constants

const (
	AppName                   = "main"
	AppNameImagePullSecret    = "image-pull-secret"
	AppNameCustomDomain       = "custom-domain"
	VmProxyAppName            = "vm-proxy"
	StorageManagerNamePrefix  = "system"
	StorageManagerAppName     = "storage-manager"
	StorageManagerAppNameAuth = "storage-manager-auth"
	WildcardCertSecretName    = "wildcard-cert"
)

func WithImagePullSecretSuffix(appName string) string {
	return appName + "-" + AppNameImagePullSecret
}

func WithCustomDomainSuffix(appName string) string {
	return appName + "-" + AppNameCustomDomain
}
