package constants

const (
	AppName                   = "main"
	AppNameImagePullSecret    = "image-pull-secret"
	AppNameCustomDomain       = "custom-domain"
	VmProxyAppName            = "vm-proxy"
	StorageManagerNamePrefix  = "system"
	StorageManagerAppName     = "storage-manager"
	StorageManagerAppNameAuth = "storage-manager-auth"
)

func ImagePullSecretSuffix(appName string) string {
	return appName + "-" + AppNameImagePullSecret
}

func CustomDomainSuffix(appName string) string {
	return appName + "-" + AppNameCustomDomain
}
