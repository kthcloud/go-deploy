package keys

const (
	// ManifestLabelName is the label name for the `name` of a manifest.
	ManifestLabelName = "app.kubernetes.io/deploy-name"
	// ManifestCreationTimestamp is the label name for the `creation timestamp` of a manifest.
	ManifestCreationTimestamp = "app.kubernetes.io/deploy-created-at"

	// K8sAnnotationClusterIssuer is the annotation name for the `cluster issuer` in a cert-manager manifest.
	K8sAnnotationClusterIssuer = "cert-manager.io/cluster-issuer"
	// K8sAnnotationCommonName is the annotation name for the `common name` in a cert-manager manifest.
	K8sAnnotationCommonName = "cert-manager.io/common-name"
	// K8sAnnotationAcmeChallengeType is the annotation name for the `acme challenge type` in a cert-manager manifest.
	K8sAnnotationAcmeChallengeType = "cert-manager.io/acme-challenge-type"
)
