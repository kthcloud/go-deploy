package constants

const (
	// AppName is the name of the main app
	AppName = "main"
	// AppNameImagePullSecret is the name of the image pull secret app in various contexts
	AppNameImagePullSecret = "image-pull-secret"
	// AppNameCustomDomain is the name of the custom domain app in various contexts
	AppNameCustomDomain = "custom-domain"

	// VmProxyAppName is the name of the VM proxy app in various contexts
	VmProxyAppName = "vm-proxy"
	// VmRootDiskName is the name of the root disk in various contexts
	VmRootDiskName = "root-disk"
	// VmParentName is the name of the parent VM in various contexts
	VmParentName = "parent"

	// SmNamePrefix is the prefix for the storage manager app in various contexts
	SmNamePrefix = "system"
	// SmAppName is the name of the storage manager app in various contexts
	SmAppName = "sm"
	// SmAppNameAuth is the name of the storage manager auth app in various contexts
	SmAppNameAuth = "sm-auth"

	// WildcardCertSecretName is the name of the wildcard cert secret in various contexts
	WildcardCertSecretName = "wildcard-cert"
)

// WithImagePullSecretSuffix returns the image pull secret app name with the given suffix
func WithImagePullSecretSuffix(appName string) string {
	return appName + "-" + AppNameImagePullSecret
}

// WithCustomDomainSuffix returns the custom domain app name with the given suffix
func WithCustomDomainSuffix(appName string) string {
	return appName + "-" + AppNameCustomDomain
}
