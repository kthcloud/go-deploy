package constants

const (
	AppName                   = "main"
	AppNameImagePullSecret    = "image-pull-secret"
	AppNameCustomDomain       = "custom-domain"
	StorageManagerNamePrefix  = "system"
	StorageManagerAppName     = "storage-manager"
	StorageManagerAppNameAuth = "storage-manager-auth"
)

func ImagePullSecretSuffix(appName string) string {
	return appName + "-image-pull-secret"
}
