package keys

const (
	// LabelDeployName is the label name for the `name` of a manifest.
	LabelDeployName = "app.kubernetes.io/deploy-name"

	// AnnotationExternalIP is the label name for the `external IP` of a manifest.
	// Right now this is only used for MetalLB manifests.
	AnnotationExternalIP = "metallb.universe.tf/loadBalancerIPs"
	// AnnotationSharedIP is the label name for the `shared IP` of a manifest.
	// Right now this is only used for MetalLB manifests.
	AnnotationSharedIP = "metallb.universe.tf/allow-shared-ip"
	// AnnotationCreationTimestamp is the label name for the `creation timestamp` of a manifest.
	AnnotationCreationTimestamp = "app.kubernetes.io/deploy-created-at"
	// AnnotationClusterIssuer is the annotation name for the `cluster issuer` in a cert-manager manifest.
	AnnotationClusterIssuer = "cert-manager.io/cluster-issuer"
	// AnnotationCommonName is the annotation name for the `common name` in a cert-manager manifest.
	AnnotationCommonName = "cert-manager.io/common-name"
	// AnnotationAcmeChallengeType is the annotation name for the `acme challenge type` in a cert-manager manifest.
	AnnotationAcmeChallengeType = "cert-manager.io/acme-challenge-type"
)
