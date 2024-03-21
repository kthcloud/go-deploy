package model

const (
	// CustomDomainStatusPending is the status of a custom domain that is pending verification.
	CustomDomainStatusPending = "pending"
	// CustomDomainStatusVerificationFailed is the status of a custom domain that failed verification.
	// This is either caused by the DNS record not being set or the DNS record not being propagated yet.
	CustomDomainStatusVerificationFailed = "verificationFailed"
	// CustomDomainStatusActive is the status of a custom domain that is active,
	// i.e., the DNS record is set and propagated.
	CustomDomainStatusActive = "active"
)

type CustomDomain struct {
	Domain string `bson:"domain"`
	Secret string `bson:"secret"`
	Status string `bson:"status"`
}
