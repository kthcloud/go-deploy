package keys

const (
	ManifestLabelName              = "app.kubernetes.io/name"
	ManifestLabelID                = "app.kubernetes.io/deploy-id"
	ManifestCreationTimestamp      = "app.kubernetes.io/deploy-created-at"
	K8sAnnotationClusterIssuer     = "cert-manager.io/cluster-issuer"
	K8sAnnotationCommonName        = "cert-manager.io/common-name"
	K8sAnnotationAcmeChallengeType = "cert-manager.io/acme-challenge-type"
)
